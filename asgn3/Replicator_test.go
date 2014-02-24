package replicator

import (
	"fmt"
	"testing" //import go package for testing related functionality
	"time"
	//      "sync"
)

func Test_DetachMoreThanHalf(t *testing.T) {
        total := 5
        var servers [5]*Replicator
        for i := 0; i < total; i++ {
                servers[i] = GetNew("config.json", i+1)
        }
        lead := -1
	ex:=-1
        for i := 0; i < 22; i++ { //No of iteration
                flag := false
                for j := 0; j < total; j++ {
                        if servers[j].IsLeader() == true && servers[j].IsDetached()==false {
                                if flag == true {
                                        t.Error("Multiple Leader")
                                        fmt.Println(">>>>>>>>> Also Leader is :", (j + 1), " for term: ", servers[j].Term())
                                } else {
                                        flag = true
                                        fmt.Println(">>>>>>>>>>Leader is :", (j + 1), " for term: ", servers[j].Term())
                                        lead = j + 1
                                        //      servers[j].Detach()
                                        //      time.Sleep(500 * time.Millisecond)
                                }
                        }
                }
		if(flag ==false){
			fmt.Println("No leader currently available")
		}
                if i == 4 {
                        fmt.Println("**********Now I will detach leader*********")
                        (&servers[lead-1]).Detach()// Detaching the LEADER itself
			ex=lead-1
                }
                if i==8 {
			fmt.Println("************ I am reattaching our Ex- Leader, i.e. server",ex+1)
                        (&servers[ex]).Attach()
                }
		if i==11{
			fmt.Println("****** Now I will detach 3 servers: 1,2,3")
			(&servers[0]).Detach()
			(&servers[1]).Detach()
			(&servers[2]).Detach()
		}
		if i==16 {
                        fmt.Println("************ I am reattaching server 2")
                        (&servers[1]).Attach()
                } 
		if i==18{
			fmt.Println("************ Attaching server 1*************")
			(&servers[0]).Attach()
		}
                if flag == false {
                        //              t.Error("No Leader")
                }
                fmt.Println("It", i, " end =========================")
                time.Sleep(500 * time.Millisecond)
        }
}


/*
func Test_General(t *testing.T) {
	total := 5
	var servers [5]*Replicator
	for i := 0; i < total; i++ {
		servers[i] = GetNew("config.json", i+1)
	}
	lead := -1
	for i := 0; i < 14; i++ { //No of iteration
		flag := false
		for j := 0; j < total; j++ {
			if servers[j].IsLeader() == true && servers[j].IsDetached()==false {
				if flag == true {
					t.Error("Multiple Leader")
					fmt.Println(">>>>>>>>> Also Leader is :", (j + 1), " for term: ", servers[j].Term())
				} else {
					flag = true
					fmt.Println(">>>>>>>>>>Leader is :", (j + 1), " for term: ", servers[j].Term())
					lead = j + 1
					//	servers[j].Detach()
					//	time.Sleep(500 * time.Millisecond)
				}
			}
		}
		if i == 4 {
			fmt.Println("**********Now I will detach leader*********")
			(&servers[lead-1]).Detach()
//			(&servers[lead-2]).Detach()
		}
		if i==8 {
			(&servers[3]).Attach()
		}
		if flag == false {
			//		t.Error("No Leader")
		}
		fmt.Println("It", i, " end =========================")
		time.Sleep(500 * time.Millisecond)
	}
}*/
