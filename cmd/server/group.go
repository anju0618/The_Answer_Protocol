// GROUPコマンド(CREATE/INVITE/JOIN/LEAVE)によるパーティ管理。
// CHAT GROUPの宛先解決にも使われる(chat.go参照)。
package main

import (
	"fmt"
	"net"
	"strings"
)

// Group はパーティ1つぶんの状態。Leaderはメンバーが0人になるか、
// Leader自身が抜けた時に再選出される(現メンバーの中で名前が
// 辞書順最小の者、removeGroupMemberLockedを参照)。
type Group struct {
	ID      string
	Leader  string
	Members map[string]struct{}
	Invited map[string]struct{}
}

// newGroup は creator をリーダー兼唯一のメンバーとする新しいGroupを作る。
func newGroup(id, creator string) *Group {
	return &Group{
		ID:      id,
		Leader:  creator,
		Members: map[string]struct{}{creator: {}},
		Invited: make(map[string]struct{}),
	}
}

// handleGroup はGROUPコマンドのサブコマンド(CREATE/INVITE/JOIN/LEAVE)を
// 引数チェックした上で、対応するhandleGroupXxxへ振り分ける。
func handleGroup(s *Server, conn net.Conn, name *string, parts []string) bool {
	if !requireArgs(conn, parts, 2) {
		return false
	}
	action := strings.ToUpper(parts[1])
	switch action {
	case "CREATE", "LEAVE":
		if !requireExactArgs(conn, parts, 2) {
			return false
		}
	case "INVITE", "JOIN":
		if !requireExactArgs(conn, parts, 3) {
			return false
		}
	default:
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if *name == "" {
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}

	switch action {
	case "CREATE":
		return handleGroupCreate(s, conn, *name)
	case "INVITE":
		return handleGroupInvite(s, conn, *name, parts[2])
	case "JOIN":
		return handleGroupJoin(s, conn, *name, parts[2])
	default:
		return handleGroupLeave(s, conn, *name)
	}
}

// handleGroupCreate はGROUP CREATEを処理する。nameを唯一のメンバー兼
// リーダーとする新しいグループを作る。既にどこかのグループに入って
// いればERR 402を返す。
func handleGroupCreate(s *Server, conn net.Conn, name string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[name]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	id := fmt.Sprintf("group.%d", s.nextGroupID+1)
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		s.nextGroupID++
		s.groups[id] = newGroup(id, name)
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleGroupInvite はGROUP INVITE <target>を処理する。nameが所属する
// グループへtargetを招待し、targetにEVT GROUP INVITEを通知する。
// target が既に何らかのグループに所属している、あるいはオンラインで
// なければエラーを返す。
func handleGroupInvite(s *Server, conn net.Conn, name, target string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	id, member := s.groupByPlayer[name]
	group := s.groups[id]
	if !member {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
		return false
	}
	if group == nil || s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[target]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	invitee := s.clients[target]
	if s.players[target] == nil || invitee == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	response, err := client.enqueueResponse("OK")
	if err == nil {
		group.Invited[target] = struct{}{}
		invitee.enqueueEvent("EVT GROUP INVITE " + group.Leader)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleGroupJoin はGROUP JOIN <leader>を処理する。nameがleaderの
// グループへの招待を受けていれば加入し、グループ全員にEVT GROUP JOINを
// ブロードキャストする。招待されていない/既に他グループに所属済みなら
// エラーを返す。
func handleGroupJoin(s *Server, conn net.Conn, name, leader string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if s.players[name] == nil {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 500 STATE_ERROR")
		return false
	}
	if _, exists := s.groupByPlayer[name]; exists {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 402 ALREADY_IN_GROUP")
		return false
	}
	id := s.groupByPlayer[leader]
	group := s.groups[id]
	if group == nil || group.Leader != leader {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	if _, invited := group.Invited[name]; !invited {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return false
	}
	response, err := client.enqueueResponse("OK group=" + id)
	if err == nil {
		group.Members[name] = struct{}{}
		s.groupByPlayer[name] = id
		s.clearGroupInvitesLocked(name)
		s.broadcastGroupLocked(group, "EVT GROUP JOIN "+name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// handleGroupLeave はGROUP LEAVEを処理する。nameが所属グループから抜ける。
func handleGroupLeave(s *Server, conn net.Conn, name string) bool {
	client := conn.(*serverClient)
	s.mu.Lock()
	if _, member := s.groupByPlayer[name]; !member {
		s.mu.Unlock()
		fmt.Fprintln(conn, "ERR 401 NOT_IN_GROUP")
		return false
	}
	response, err := client.enqueueResponse("OK")
	if err == nil {
		s.removeGroupMemberLocked(name)
	}
	s.mu.Unlock()
	if err != nil {
		return true
	}
	return client.waitResponse(response) != nil
}

// broadcastGroupLocked は group の全メンバー(オンラインの者のみ)に
// event を送る。
func (s *Server) broadcastGroupLocked(group *Group, event string) {
	for member := range group.Members {
		if client := s.clients[member]; client != nil {
			client.enqueueEvent(event)
		}
	}
}

// clearGroupInvitesLocked は name 宛の招待を全グループから取り消す
// (グループ作成時・グループ加入時に呼ばれる)。
func (s *Server) clearGroupInvitesLocked(name string) {
	for _, group := range s.groups {
		delete(group.Invited, name)
	}
}

// removeGroupMemberLocked は name をそのグループから除去する(GROUP LEAVE・
// 切断・退出時に呼ばれる)。グループが空になれば削除し、抜けたのが
// リーダーなら残りメンバーの中で名前が辞書順最小の者を新リーダーに
// する。残りメンバーへEVT GROUP LEAVEを通知する。
func (s *Server) removeGroupMemberLocked(name string) {
	id, member := s.groupByPlayer[name]
	if !member {
		return
	}
	delete(s.groupByPlayer, name)
	group := s.groups[id]
	if group == nil {
		return
	}
	delete(group.Members, name)
	if len(group.Members) == 0 {
		delete(s.groups, id)
		return
	}
	if group.Leader == name {
		group.Leader = ""
		for member := range group.Members {
			if group.Leader == "" || member < group.Leader {
				group.Leader = member
			}
		}
		clear(group.Invited)
	}
	s.broadcastGroupLocked(group, "EVT GROUP LEAVE "+name)
}
