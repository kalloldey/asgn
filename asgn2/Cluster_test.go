package cluster

import (
	//	"fmt"
	"strconv"
	"strings"
	"testing" //import go package for testing related functionality
	"time"
	//	"sync"
)

func sendToServers(servers [150]*Raftserver) {
	total := 150
	for i := 0; i < total; i++ {
		smsg := string("Server|" + strconv.Itoa(i+1) + "|says Hi") //Do not change the format of the msg ..
		servers[i].Outbox() <- &Envelope{Pid: -1, MsgId: 0, Msg: smsg}
		//fmt.Println("Send msg for server :",i+1)
		time.Sleep(800 * time.Millisecond)
	}
}

/*
Test Case Description:
To check what happend when huge number servers simultaneously broadcasts messages. Here 150 server
started and they broadcast to all.
*/
func Test_OneHundredFiftyServers(t *testing.T) {
	total := 150
	var servers [150]*Raftserver
	//        var Fin        sync.WaitGroup //dummy .. workaround ..
	//      Fin.Add(1)
	for i := 0; i < total; i++ {
		servers[i] = New("config150.json", i+1)
	}
	totMsg := 0
	cumulativeSum := 0
	go sendToServers(servers)
	for i := 0; i < total; i++ {
		for j := 0; j < total; j++ { //you will get total-1 msg ..
			//      fmt.Println("Waiting for channelof ",j)
			if i != j {
				opt := <-servers[j].Inbox()
				temp := strings.Split(opt.Msg, "|")
				//                                fmt.Println("i=",i,"Sender: ", temp[1], "Receiver: ", j+1)
				senderPid, _ := strconv.Atoi(temp[1])
				if senderPid == i {
					t.Error("Received own message !!")
				}
				totMsg++
				cumulativeSum++
			}
		}
		if totMsg != (total - 1) {
			t.Error("Message dropped in between")
		}
		totMsg = 0
	}
	//Total cumulative message is 150* 149 =22350
	if cumulativeSum != 22350 {
		t.Error("Total message is not same as expected")
	}
	//        Fin.Wait()
}
/*
Test Description: Test with HUGE size message. The goal is to check if the huge size
message is delivered properly. This is done in Every server send to its next server in
this manner with last server send to first server. Circle manner precisely.
Message Size: 37888 bytes
*/
/*
func Test_HugeMessagTenServer(t *testing.T) {
	total := 10
	var servers [10]*Raftserver
	for i := 0; i < total; i++ {
		servers[i] = New("config10.json", i+1)
	}
	hugeMsg := "abcdef1238910212320**23jdfm@@89200!!!"

	for i := 0; i < 10; i++ {
		hugeMsg = string(hugeMsg + hugeMsg)
	}

	for i := 0; i < total; i++ {
		time.Sleep(2 * time.Millisecond)
		willSendTo := (i+1)%total + 1
		servers[i].Outbox() <- &Envelope{Pid: willSendTo, MsgId: 0, Msg: hugeMsg}
		//fmt.Println("Send msg for server :",i+1)
	}
	for i := 0; i < total; i++ {
		opt := <-servers[i].Inbox()
		if opt.Msg != hugeMsg {
			t.Error("Not Same")
		}
	}
}*/
