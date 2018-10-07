package lib

import (
	"fmt"
)

func statusMapOfStatusVector(x []PeerStatus) map[string]uint32 {
	m := make(map[string]uint32)
	for _, peer_status := range x {
		m[peer_status.Identifier] = peer_status.NextID
	}
	return m
}

var Status_Self_Knows_More = 1
var Status_Equal = 0
var Status_Remote_Knows_More = -1

func util_compare(m_self map[string]uint32, m_remote map[string]uint32) (int, *PeerStatus) {
	for name, id_self := range m_self {
		if id_remote, ok := m_remote[name]; ok {
			if id_remote < id_self {
				// I have informations about name that remote
				// doesn't have: I send it to him
				return Status_Self_Knows_More, &PeerStatus{Identifier: name, NextID: id_remote - 1}
			} else if id_remote > id_self {
				// Remote has more knowledge than me
				return Status_Remote_Knows_More, &PeerStatus{Identifier: name, NextID: id_self - 1}
			}
		} else {
			// I have more knowledge than the other as I know a peer he doesn't
			return Status_Self_Knows_More, &PeerStatus{Identifier: name, NextID: id_self - 1}

		}
	}
	return Status_Equal, nil
}

func print(self []PeerStatus) string {
	out := ""
	for _, x := range self {
		out += x.Identifier + ": " + fmt.Sprint(x.NextID) + "; "
	}
	return out
}

/*
Compare two vector status
Returns 0 if they are both equal
-1 if the remote has a status message that self doesn't have
1 otherwise. In this case, also returns the corresponding PeerStatus.
*/
func CompareStatusVector(self []PeerStatus, remote []PeerStatus) (int, *PeerStatus) {
	m_self := statusMapOfStatusVector(self)
	m_remote := statusMapOfStatusVector(remote)

	eta, x := util_compare(m_self, m_remote)
	if eta == Status_Equal {
		eta2, x2 := util_compare(m_remote, m_self)
		//		log.Printf("%v <=> %v = %v\n", print(self), print(remote), -eta2)
		return -eta2, x2
	} else {
		//		log.Printf("%v <=> %v = %v\n", print(self), print(remote), eta)
		return eta, x
	}
	//	log.Printf("%v <=> %v = %v\n", print(self), print(remote), 0)

	return Status_Equal, nil
}
