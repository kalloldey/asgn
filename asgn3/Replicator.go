package replicator

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	//	"../cluster/"
	//	"/home/kallol/Prog/GO/src/github.com/kalloldey/assignment/cluster"
	"encoding/json"
)

func (r Replicator) P(a string, b int, c int){
	ck:=1
	if(ck==1){
		fmt.Println("[(",r.MyPid ,") ",a," || ",b," :: ",c," ]")
	}
}

type Raft interface {
	Term() int
	IsLeader() bool
	Detach()
	Attach()
	IsDetached() bool
}

type Replicator struct {
	MyPid        int
	LeaderFlag   int //0: I am follower, 1: I am candidate, 2: I am leader
	CurrentTerm  int
	TotalPeer    int
	PidOfLeader  int //-1 when do not know
	Detached     int //set to 1 to detach this server from others ...
	BackServer   *Raftserver
	TimeOutMin   int
	TimeOutRand  int
	HBRecChan    chan int
	VoteReceived int
	BallotReceived int
	Locker       *sync.Mutex
	LeadLock     *sync.Mutex
}

var sett struct {
	SelfHandle  string `json:"selfHandle"`
	PeersPid    string `json:"peersPid"`
	PeersHandle string `json:"peersHandle"`
	StartAddr   int    `json:"startAddress"`
	StartMsgId  int    `json:"startMsgId"`
	BufferSize  int    `json:"bufSize"`
	TimeOutMin  int    `json:"timeoutMin"`  //Minimum timeout length
	TimeOutRand int    `json:"timeoutRand"` //One value in between 0 to TimeOutRand will be choosen and added to TimtOutMin
}

func (r Replicator) Term() int {
	return r.CurrentTerm
}
func (r Replicator) IsLeader() bool {
	if r.LeaderFlag == 2 {
		return true
	}
	return false
}

func (r Replicator) SetLeadFlag(a int) {
	//	r.LeadLock.Lock()
	r.LeaderFlag = a
	//	r.LeadLock.Unlock()
}

func (r *Replicator) Detach() {
	r.P("I am inside Detach call",0,0)
	r.Detached = 1
}

func (r *Replicator) Attach() {
	r.P("I am inside Attach call",0,0)
	r.CurrentTerm = GetTerm(r)
	r.Detached = 0
}

func (r Replicator) IsDetached() bool {
	if r.Detached == 1 {
		return true
	}
	return false
}

func GetTerm(rp *Replicator) int {
	myPid := rp.MyPid
	tmflname := string("TERMDB_" + strconv.Itoa(myPid))
	termfile, err := os.Open(tmflname)
	myTerm := 0
	if err != nil { //If no file exist then create one
		myTerm = 0
		termfile, err = os.Create(tmflname)
		//              fmt.Println("If part")
		d3 := []byte(string(strconv.Itoa(myTerm) + "\n"))
		ioutil.WriteFile(tmflname, d3, 0644)
	}
	defer termfile.Close()
	bt := make([]byte, 10)
	termfile.Read(bt)
	lines := strings.Split(string(bt), "\n")
	kl := lines[0]
	//        fmt.Println("String read : ",kl)
	myTerm, err = strconv.Atoi(kl)
	if err != nil {
		fmt.Println("Err on conv")
	}
	return myTerm
} //OK

var mutex = &sync.Mutex{}

func SetTerm(rp *Replicator, newTerm int) int {
	mutex.Lock()
	tmflname := string("TERMDB_" + strconv.Itoa(rp.MyPid))
	termfile, err := os.Open(tmflname)
	if err != nil {
		termfile, err = os.Create(tmflname)
	}
	defer termfile.Close()
	rp.CurrentTerm = newTerm
	d3 := []byte(string(strconv.Itoa(newTerm) + "\n"))
	ioutil.WriteFile(tmflname, d3, 0644)
	mutex.Unlock()
	return newTerm
} //OK

func StartVote(rp *Replicator) {
	//Increase term by one
	if rp.LeaderFlag ==2 {  //I am already leader .. no more election needed
		return
	}
	if rp.Detached == 1 {//I am detached why tell me to start vote ???
		rp.P("I am detached",0,0)
		return 
	}
	voteForTerm := rp.CurrentTerm + 1
	rp.P("Vote starting for term: ",voteForTerm,0)	
	rp.VoteReceived = 1      //will be only valid when I am in candidate state
	rp.BallotReceived = 1
	rp.LeaderFlag = 1        //convert to candidate state
	SetTerm(rp, voteForTerm) //Save in file
	//Vote for me msg pattern:    VOTEME$<MyPid>$ForTheTerm$
	sm := string("VOTEME$" + strconv.Itoa(rp.MyPid) + "$" + strconv.Itoa(voteForTerm))
//	rp.Locker.Lock()
	rp.P("SendMsg= "+sm,0,0)
	rp.BackServer.Outbox() <- &Envelope{Pid: -1, MsgId: 0, Msg: sm}
//	rp.Locker.Unlock()
}

func ElectionCommison(rp *Replicator) {  //Detached case taken care by the StartVote .. so here can ignore
	for {
	/*	if rp.Detached == 1 {
			continue
		}
		if rp.LeaderFlag == 2 {
			continue
		}*/
		tmout := time.Duration(rp.TimeOutMin + rand.Intn(rp.TimeOutRand+rp.MyPid*2))
		select {
		case <-rp.HBRecChan:
			//HB received go for the next iteration of for loop ..
			//NO OP
			continue
		case <-time.After(tmout * time.Millisecond):
			StartVote(rp)
		}
	}
}


//func (rp Replicator) SendHBM(){  //<--abandoned

func SendHBM(rp *Replicator) { //when ever i am leader send Heart beat message  //Detached case taken care
	for {
		time.Sleep(time.Duration(100) * time.Millisecond)
		if rp.IsLeader() == false {
			continue
		}
		if rp.Detached == 1 {
			rp.P("I am detached",0,0)
			time.Sleep(time.Duration(100) * time.Millisecond)  //if I am detached .. so use this to slow down checking
			continue
		}
		sm := string("HEARTBEAT$" + strconv.Itoa(rp.MyPid) + "$" + strconv.Itoa(rp.CurrentTerm))
//		rp.Locker.Lock()
		rp.P("Send HBM for term, detach value ",rp.CurrentTerm,rp.Detached)
		rp.BackServer.Outbox() <- &Envelope{Pid: -1, MsgId: 0, Msg: sm}
//		rp.Locker.Unlock()
	}
}

//Vote for me msg pattern:    VOTEME$<MyPid>$ForTheTerm$
//Reply with Deny Vote msg pattern:   VOTEDENY $ Pid of server $ For the term $
//Reply with Grant Vote msg pattern:   VOTEGRANT $Pid of server $ For the term $
//Leader Heartbeat msg pattern: HEARTBEAT$Pid of the leader $for the term $

func TelecomMinistry(rp *Replicator) {
	for {
		rec := <-rp.BackServer.Inbox()
//		rp.P("Detach valu of mine =",rp.Detached,0)
		if rp.Detached == 1 {
			rp.P("I am detached",0,0)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}
		temp := strings.Split(rec.Msg, "$")
		//		fmt.Println("Msg found:  ",rec.Msg)
		serverPid, _ := strconv.Atoi(temp[1])
		//		fmt.Println("Temp1: ",temp[1]," | serverPid: ",serverPid)
		forTerm, _ := strconv.Atoi(temp[2])

		switch temp[0] {
		case "VOTEDENY":
			{
				if forTerm > rp.CurrentTerm {
					rp.P("Get deny msg from,term",serverPid,forTerm)
					rp.LeaderFlag = 0 //become follower again as big term is present in somewhere
					SetTerm(rp, forTerm)
					//Doubt to clear: Will VOTEDENY will get the current update term from the responder server ??
				}
			}
		case "VOTEME":
			{
				serverPid, _ := strconv.Atoi(temp[1])
				forTerm, _ := strconv.Atoi(temp[2])
				if rp.CurrentTerm < forTerm {
					//covert own to follower state rp.
					SetTerm(rp, forTerm)
					denmsg := string("VOTEGRANT$" + strconv.Itoa(rp.MyPid) + "$" + strconv.Itoa(forTerm))
//					rp.Locker.Lock()
					rp.LeaderFlag = 0
					rp.P("I am GIVING vote to, in term",serverPid,forTerm)
					rp.BackServer.Outbox() <- &Envelope{Pid: serverPid, MsgId: 0, Msg: denmsg}
//					rp.Locker.Unlock()
				} else {
					denmsg := string("VOTEDENY$" + strconv.Itoa(rp.MyPid) + "$" + strconv.Itoa(forTerm))
//					rp.Locker.Lock()
					rp.P("I am DENYing vote to,term ",serverPid,forTerm)
					rp.BackServer.Outbox() <- &Envelope{Pid: serverPid, MsgId: 0, Msg: denmsg}
//					rp.Locker.Unlock()
				}
			}
		case "HEARTBEAT":
			{
				if rp.CurrentTerm > forTerm {
					//Not possible error case
					/*This only can happen if a node goes down when it was leader and after sometime
					came back and continue to send hbm .. however this will be automatically taken 
					care by that node and it will steps down to follower state once the actual leader 
					send message*/
				} else {
					if(forTerm !=rp.CurrentTerm){
						SetTerm(rp, forTerm)
					}
					rp.P("Get HBM from leader, term", serverPid, forTerm)
					rp.HBRecChan <- 1
					rp.LeaderFlag = 0
					rp.PidOfLeader = serverPid
				}

			}
		case "VOTEGRANT":
			{
				if rp.LeaderFlag == 1 && forTerm == rp.CurrentTerm {
					rp.VoteReceived++					
					rp.P("Received VOTE,from in term",serverPid,forTerm)					
					if rp.VoteReceived >= ((rp.TotalPeer / 2) + (rp.TotalPeer % 2)) { //self become leader
						rp.LeaderFlag = 2
						rp.P("I am now leader,total vote rec,term",rp.VoteReceived,forTerm)
						rp.PidOfLeader = rp.MyPid
					}
				}
			}
		default:
			{
				//NO OP
			}
		}

	}
}

func GetNew(FileName string, PidArg int) *Replicator {
	//File read starts ....
//	fmt.Println("Start of GetNew")
	repl := new(Replicator)
	configFile, err := os.Open(FileName)
	if err != nil {
		fmt.Println("opening config file: ", FileName, "..", err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&sett); err != nil {
		fmt.Println("Error in parsing config file ", err.Error())
	}

	repl.MyPid = PidArg
	repl.LeaderFlag = 0
	repl.PidOfLeader = -1
	//termFile, err := os.Open(string("termfileDB" + strconv.Itoa(repl.MyPid))) //termfileDB1
	repl.HBRecChan = make(chan int, 1)
	repl.CurrentTerm = GetTerm(repl)
	repl.TimeOutMin = sett.TimeOutMin
	repl.TimeOutRand = sett.TimeOutRand
	repl.TotalPeer = len(strings.Split(sett.PeersPid, ","))
	repl.P("Total Peer : ",repl.TotalPeer,0)
	repl.Detached = 0
	repl.Locker = &sync.Mutex{}
	repl.LeadLock = &sync.Mutex{}
	/*	New_DirectArg(self pid, "pid of peers in a comma separated string", Strat address of the port as int, Self handle in string, Peers handle in string)*/
	repl.BackServer = New_DirectArg(repl.MyPid, sett.PeersPid, sett.StartAddr, sett.SelfHandle, sett.PeersHandle)
	go TelecomMinistry(repl)
	go ElectionCommison(repl)
	go SendHBM(repl)
	repl.P("Initiate ",0,0)
//	fmt.Println("End of GetNEw .. returing")
	return repl
}
