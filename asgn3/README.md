Leader Election Algorithm Implementation
========================================
What is this?
--------------
The overall goal of the project is to implement the Raft consensus algorithm and desiging a distributed
key value store system. Leader election and keep control of Leader in a particular term is a important
part of Raft. This assignment design Leader election portion of Raft paper.

How this package is coming into the picture?
--------------------------------------------
This package having Replicator is a partial implementation of the final Replicator object. Currently the 
available features are:

- Can elect leader
- Can detect when no leader is present
- Leaders can send heart beat message
- Participate in election either as candidate or follower


What is usefullness of this package?
------------------------------------
This package is one of the most integral part of Raft implementation. Without a leader in a particular term
the normal functioning will not be possible. 



How to use it?
-------------
*Please refer to general guidelines for importing a github package. Import this package into your worksapce*

- Replicator object can be created using a call to GetNew("config file name", "pid of the Replicator server")
- Once the call return the object, it is assured that all the functionings are started
- You can induce illusion of partitioning to your server by calling Detach() and later can attach the server
again by calling Attach()

What is the config file format?
------------------
Configuration file need to be in json
Below is a example of config.json file
	{

        "selfHandle":"tcp://127.0.0.1:", 

        "peersPid": "1,2",

        "peersHandle":"tcp://127.0.0.1:",

        "startAddress":5000,

        "startMsgId":1000,

        "bufSize":5000,

	"timeOutMin":100,

	"timeOutRand":200

	}

Here handle means the communication end point for a cluster.
Please make sure the Pid given to instantiate a server exist in the peersPid list,
for more clarification I can say, peersPid will have all the servers Pid including 
the server which will use the config file

What are the technology used?
----------------------------
- **Language:** GO 
- **Library:** zmq V4 was used
- **Paper Follwed:** Raft

Who are the people associated with this work?
---------------------------------------------
**Kallol Dey**, a senior post graduate student  of CSE IIT Bombay has implemented this.
Specification, idea and overall guidance was provided by **Dr. Sriram Srinivasan**.


Current status of the project?
------------------------------
It was tested with several configuration. 
For basic 5 servers it is successful in electing leader and seding heart beat message
For particioning of 1 server it, need some debug to handle
More partitioning also need some debug work, but it will be perfect very soon
