Clusterring Services
====================
What is this?
--------------
In distributed environment several time you will feel the requirement of implementing some consensus 
algorithm to deal with several issues that are very common in distributed system. As this issues
and corresponding consensus algorithms are very easily available in several books and papers, I will
not describe this here. When implementing consensus algorithm or any other algorithm in distributed
system you will need several clusters. The clusters will not be stand alone and they will have the
ability to cimmunicate with each other. My package is one kind of implementation of clusters.


How this package is coming into the picture?
--------------------------------------------
This package having name cluster is a full fledged implementation of cluster model. The clusters have
the below facilities:

- Can send message to other peers
- Can receive message from other peers
- Can add a new peer
- Delete a peer


What is usefullness of this package?
------------------------------------
This package can be used to make your own clusters. You can send and receive message to and from the
peer clusters. With this underlaying layer any consensus algorithm can be implemented over it.



How to use it?
-------------
*Please refer to general guidelines for importing a github package. Import this package into your worksapce*
- Cluster need to be initated by making a call to New(). New will take two parameters, configuration file
in json format and if of individual cluster. You can keep the config file same accross clusters.
- Once you import this package you can use the Inbox() channel to receive message
- Use the Outbox() method to send message to other peers
- DelPeer and AddPeer functionality is also there
- Outbox and Inbox both will return channels so user can observe the channels for possible presence of data

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
        "bufSize":5000
	}
Here handle means the communication end point for a cluster.
Please make sure the Pid given to instantiate a server exist in the peersPid list,
for more clarification I can say, peersPid will have all the servers Pid including 
the server which will use the config file

What are the technology used?
----------------------------
- **Language:** GO 
- **Library:** zmq V4 was used


Who are the people associated with this work?
---------------------------------------------
**Kallol Dey**, a senior post graduate student  of CSE IIT Bombay has implemented this.
Specification, idea and overall guidance was provided by **Dr. Sriram Srinivasan**.


Current status of the project?
------------------------------
It was tested with several configuration. For 10 server it was working perfectly.
For 700 server it is gving error like, system can not open this number of file. 
For 200 server inbetween the test is halting, need to analys this properly.

