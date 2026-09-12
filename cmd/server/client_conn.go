// 1接続ぶんの非同期送信キュー(serverClient)。各ハンドラはServer.muを
// 保持したまま直接ソケットに書き込むことはせず、ここにメッセージを
// enqueueするだけにする。実際の書き込みは専用のwriteLoop goroutineが
// 行うので、ネットワークI/Oの遅延で全プレイヤー共有のロックを塞ぐことがない
// (README「Architecture」の「Non-blocking broadcast」参照)。
package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var errSendQueueFull = errors.New("client send queue full")

const defaultWriteTimeout = 10 * time.Second

// outboundMessage はwriteLoopに渡す1件の送信データ。done が非nilなら、
// 書き込み結果(waitResponseで待てる)を通知する。
type outboundMessage struct {
	data []byte
	done chan writeResult
}

// writeResult は実際にソケットへ書き込んだ結果。
type writeResult struct {
	n   int
	err error
}

// serverClient はnet.Connをラップし、非同期送信キューと選択中の言語
// (locale.go参照)を持たせたもの。net.Connを埋め込んでいるので、他の
// コードからは普通のnet.Connとして(Write以外は)扱える。
type serverClient struct {
	net.Conn
	out          chan outboundMessage
	done         chan struct{}
	writeTimeout time.Duration
	closeOnce    sync.Once
	locale       string
}

// newServerClient はデフォルトの書き込みタイムアウトでserverClientを作る。
func newServerClient(conn net.Conn) *serverClient {
	return newServerClientWithWriteTimeout(conn, defaultWriteTimeout)
}

// newServerClientWithWriteTimeout はserverClientを作り、専用のwriteLoop
// goroutineを起動する(テストでタイムアウトを変えたい場合のためtimeout
// を引数に取れるようにしている)。
func newServerClientWithWriteTimeout(conn net.Conn, timeout time.Duration) *serverClient {
	client := &serverClient{
		Conn:         conn,
		out:          make(chan outboundMessage, 64),
		done:         make(chan struct{}),
		writeTimeout: timeout,
	}
	go client.writeLoop()
	return client
}

// writeLoop はclient.outから送信メッセージを取り出し、実際にソケットへ
// 書き込み続ける専用goroutine。書き込みエラーが起きたら接続をCloseして
// 終了する。newServerClientWithWriteTimeoutからgoで起動される。
func (client *serverClient) writeLoop() {
	for {
		select {
		case message := <-client.out:
			err := client.Conn.SetWriteDeadline(time.Now().Add(client.writeTimeout))
			n := 0
			if err == nil {
				n, err = client.Conn.Write(message.data)
			}
			if err == nil && n != len(message.data) {
				err = io.ErrShortWrite
			}
			if message.done != nil {
				message.done <- writeResult{n: n, err: err}
			}
			if err != nil {
				client.Close()
				return
			}
		case <-client.done:
			return
		}
	}
}

// Write はnet.Connのio.Writerインターフェースを満たすための実装。
// writeLoopにメッセージを渡し、書き込みが完了する(または接続が閉じる)
// まで同期的にブロックする。fmt.Fprintln(conn, ...)のような直接書き込み
// から呼ばれる。
func (client *serverClient) Write(data []byte) (int, error) {
	message := outboundMessage{
		data: append([]byte(nil), data...),
		done: make(chan writeResult, 1),
	}
	select {
	case client.out <- message:
	case <-client.done:
		return 0, net.ErrClosed
	}
	select {
	case result := <-message.done:
		return result.n, result.err
	case <-client.done:
		return 0, net.ErrClosed
	}
}

// enqueueResponse は1行のレスポンスを送信キューへ積む(ノンブロッキング。
// キューが満杯ならerrSendQueueFullを返す)。呼び出し側はServer.muを
// 保持したまま呼べる。返り値のチャネルをwaitResponseに渡すと、実際の
// 書き込み完了を待てる。
func (client *serverClient) enqueueResponse(line string) (<-chan writeResult, error) {
	message := outboundMessage{
		data: []byte(line + "\n"),
		done: make(chan writeResult, 1),
	}
	select {
	case client.out <- message:
		return message.done, nil
	case <-client.done:
		return nil, net.ErrClosed
	default:
		return nil, errSendQueueFull
	}
}

// waitResponse はenqueueResponseが返したチャネルから書き込み結果を
// 受け取る。Server.muを解放した後に呼ぶ(ネットワークI/Oの完了を
// ロックを持たずに待つため)。
func (client *serverClient) waitResponse(done <-chan writeResult) error {
	select {
	case result := <-done:
		return result.err
	case <-client.done:
		return net.ErrClosed
	}
}

// enqueueEvent はEVT行(非同期プッシュ通知)を送信キューへ積む。
// enqueueResponseと違って完了を待つ手段を提供しない「投げっぱなし」。
// キューが満杯の場合は接続を切ってしまう(詰まったクライアント1つが
// 他のブロードキャストを止めないようにするため)。
func (client *serverClient) enqueueEvent(line string) {
	select {
	case <-client.done:
		return
	default:
	}
	select {
	case client.out <- outboundMessage{data: []byte(line + "\n")}:
	case <-client.done:
	default:
		client.Close()
	}
}

// Close は接続を閉じる。複数回呼ばれても安全(sync.Onceで一度しか
// 実行されない)。
func (client *serverClient) Close() error {
	var err error
	client.closeOnce.Do(func() {
		close(client.done)
		err = client.Conn.Close()
	})
	return err
}
