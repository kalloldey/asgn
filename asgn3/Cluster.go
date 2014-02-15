
package replicator

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"os"
	"strconv"
	"strings"
	"sync"
)

/* Need To Support:
   Pid() int
   Peers() []int
   AddPeer(n int)
   DelPeer(n int)
   Outbox() chan *Envelope
   Inbox() chan *Envelope
*/
/* Way followed for connection: While sending msg act as client to the remote server
While want to receive packet be a server, remote client will send you packet. */

const MAX_PEER int = 1000
const ERR int = 404

//Actual server object
type Raftserver struct {
	MyPid      int
	TotalPeer  int
	PeersPid   [MAX_PEER]int
	Server     *zmq.Socket
	OwnEnd     string
	PeerHandle string
	StartAddr  int
	Fin        sync.WaitGroup
	RecChan    chan *Envelope
	//Need some handler or identification for connection purpose...
} //WR

//Object to read json file

var settings struct {
	SelfHandle  string `json:"selfHandle"`
	PeersPid    string `json:"peersPid"`
	PeersHandle string `json:"peersHandle"`
	StartAddr   int    `json:"startAddress"`
	StartMsgId  int    `json:"startMsgId"`
	BufferSize  int    `json:"bufSize"`
}

func (r Raftserver) Pid() int {
	return r.MyPid
} //FYN

func (r Raftserver) Wait() {
	r.Fin.Wait()
}
func (r Raftserver) Peers() []int {
	return r.PeersPid[0:]
} //FYN

func (r Raftserver) AddPeer(n int) {
	r.PeersPid[r.TotalPeer] = n
	r.TotalPeer = r.TotalPeer + 1
} //FYN

func (r Raftserver) DelPeer(n int) {
	if r.TotalPeer == 1 {
		r.PeersPid[0] = ERR
		r.TotalPeer = 0
	} else if r.TotalPeer == 0 {
		return
	} else {
		for i := 0; i < r.TotalPeer; i++ {
			if r.PeersPid[i] == n {
				r.PeersPid[i] = r.PeersPid[r.TotalPeer-1]
				r.TotalPeer = r.TotalPeer - 1
				break
			}
		}
	}
} //FYN

//This will directly accept the arguments
/*New_DirectArg(self pid, "pid of peers in a comma separated string", Strat address of the port as int, Self handle in string, Peers handle in string)*/
func New_DirectArg(PidArg int, ArgPeersPid string,ArgStartAddr int, ArgSelfHandle string, ArgPeersHandle string) *Raftserver { //To create the server object
	temp := strings.Split(ArgPeersPid, ",")
	rfs := new(Raftserver) //Instatiate a private server object
	j := 0
	for i := 0; i < len(temp); i++ {
		tm, _ := strconv.Atoi(temp[i])
		if tm != PidArg {
			rfs.PeersPid[j] = tm
			j++
		}
	}
	//File data read ended ....
	rfs.MyPid = PidArg
	rfs.TotalPeer = len(temp) - 1
	//	fmt.Println("[Instantiate]Total Peer: ",rfs.TotalPeer)
	//Read Peer nethandle form file
	rfs.StartAddr = ArgStartAddr
	//	fmt.Println("[New]:startaddr", rfs.StartAddr)
	rfs.PeerHandle = ArgPeersHandle
	rfs.OwnEnd = string(ArgSelfHandle + strconv.Itoa(rfs.StartAddr+rfs.MyPid)) //read from file
	//	fmt.Println("Own End point =",rfs.OwnEnd)
	//Do configure the netowork setup portion here
	rfs.Server, _ = zmq.NewSocket(zmq.PULL)
	rfs.Server.Bind(rfs.OwnEnd)
	rfs.Fin.Add(1)
	//Make the receive and send channels ..
	//	fmt.Println("Server Instantion done .. returning..")
	rfs.RecChan = make(chan *Envelope, 1)
	go rfs.recRoutine()

	return rfs
} //WR


//This will take filename as argument
func New(FileName string, PidArg int) *Raftserver { //To create the server object
	//File read starts ....
	configFile, err := os.Open(FileName)
	if err != nil {
		fmt.Println("opening config file: ", FileName, "..", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&settings); err != nil {
		fmt.Println("Error in parsing config file ", err.Error())
	}
	temp := strings.Split(settings.PeersPid, ",")
	rfs := new(Raftserver) //Instatiate a private server object
	j := 0
	for i := 0; i < len(temp); i++ {
		tm, _ := strconv.Atoi(temp[i])
		if tm != PidArg {
			rfs.PeersPid[j] = tm
			j++
		}
	}
	//File data read ended ....
	rfs.MyPid = PidArg
	rfs.TotalPeer = len(temp) - 1
	//	fmt.Println("[Instantiate]Total Peer: ",rfs.TotalPeer)
	//Read Peer nethandle form file
	rfs.StartAddr = settings.StartAddr
	//	fmt.Println("[New]:startaddr", rfs.StartAddr)
	rfs.PeerHandle = settings.PeersHandle
	rfs.OwnEnd = string(settings.SelfHandle + strconv.Itoa(rfs.StartAddr+rfs.MyPid)) //read from file
	//	fmt.Println("Own End point =",rfs.OwnEnd)
	//Do configure the netowork setup portion here
	rfs.Server, _ = zmq.NewSocket(zmq.PULL)
	rfs.Server.Bind(rfs.OwnEnd)
	rfs.Fin.Add(1)
	//Make the receive and send channels ..
	//	fmt.Println("Server Instantion done .. returning..")
	rfs.RecChan = make(chan *Envelope, 1)
	go rfs.recRoutine()

	return rfs
} //WR

func sendRoutine(msg chan *Envelope, r *Raftserver) {
	//	fmt.Println("[send Routine] start")
	//Broadcast or Unicast
	om := <-msg
	//	fmt.Println("[send:] to pid: ", om.Pid)
	sendmsg := string(strconv.Itoa(om.Pid) + "#" + strconv.Itoa(om.MsgId) + "#" + om.Msg + "#") //msg to send out
	if om.Pid == -1 {                                                                           //Do a Broadcast
		for i := 0; i < r.TotalPeer; i++ { //Improve it by calling simultaneous goroutines instead of this ordered flow
			endpoint := string(r.PeerHandle + strconv.Itoa(r.StartAddr+r.PeersPid[i]))
			client, _ := zmq.NewSocket(zmq.PUSH)
			client.Connect(endpoint)
			//			fmt.Println("[Server ]", r.MyPid,"sending msg to: ",r.PeersPid[i])
			client.SendMessage(sendmsg)
			client.Close()
		}
	} else { //Do a unicast
		endpoint := string(r.PeerHandle + strconv.Itoa(r.StartAddr+om.Pid))
		client, _ := zmq.NewSocket(zmq.PUSH)
		//                fmt.Println("[send endpoint]:  ",endpoint)
		client.Connect(endpoint)
		client.SendMessage(sendmsg)
		client.Close()
	}
	r.Fin.Done() //telling I am over ..
}

func (r Raftserver) Outbox() chan *Envelope {
	//	fmt.Println("[Outbox] start")
	getMsg := make(chan *Envelope, 1000)
	go sendRoutine(getMsg, &r)
	return getMsg
} //FYN

func (r *Raftserver) recRoutine() {
	for {
		//	fmt.Println("[Inbox]-entry befor RecvMsg")
		msg, err := r.Server.RecvMessage(0)
		if err != nil {
			//fmt.Println("Error in receiving msg ..")
		}
		//		fmt.Println("[Inbox of ",r.MyPid,"] Msg received", msg)
		//msg is a string of format Pid#MsgId#Msg
		var temp []string
		temp = strings.Split(msg[0], "#") //concat all of msg array dont just take only first
		interEnv := Envelope{}
		interEnv.Pid, _ = strconv.Atoi(temp[0])
		interEnv.MsgId, _ = strconv.Atoi(temp[1])
		interEnv.Msg = temp[2]
		r.RecChan <- &interEnv
		//		<- r.RecChan
	}
}

func (r Raftserver) Inbox() chan *Envelope {
	//	retEnv := make(chan *Envelope, 1000)
	//	go recRoutine(retEnv, &r)
	return r.RecChan
} //FYN
