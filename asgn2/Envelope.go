/*************************************************************
Author : Kallol Dey
Description: Package with Envelope format and Server interface.
This rules need to be followed by every member of the cluster
***************************************************************/

package cluster

const (
	BROADCAST = -1
)

type Envelope struct {
	Pid   int    // Pid is always set to the original sender
	MsgId int    //Unique id number for message
	Msg   string //For this time being kept as string will replace with something complex
}

func Addme(a int, b int) int {
	return a + b
}

type Server interface {
	Pid() int
	Peers() []int
	AddPeer(n int)
	DelPeer(n int)
	Outbox() chan *Envelope
	Inbox() chan *Envelope
}
