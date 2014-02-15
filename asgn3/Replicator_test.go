package replicator

import (
              "fmt"
        "testing" //import go package for testing related functionality
        "time"
        //      "sync"
)

func Test_General(t *testing.T){
	total:=5
	var servers [5]*Replicator
	for i:=0;i<total ;i++{
		servers[i]=GetNew("config.json",i+1)
	}
	for i:=0 ;i<9 ;i++{ //No of iteration
		flag:=false
		for j:=0 ;j < total;j++{
			if(servers[j].IsLeader()==true){
				if(flag == true){
//					 t.Error("Multiple Leader")
					fmt.Println(">>>>>>>>> Also Leader is :",(j+1)," for term: ",servers[j].Term())
				}else{
					flag = true
					fmt.Println(">>>>>>>>>>Leader is :",(j+1)," for term: ",servers[j].Term())
					servers[j].Detach()
					time.Sleep(500*time.Millisecond)
				}
			}
		}	
		if(flag ==false){
//			t.Error("No Leader")
		}
		fmt.Println("It",i, " end =========================")
		time.Sleep(500*time.Millisecond)
	}
}
