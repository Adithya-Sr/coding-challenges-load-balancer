package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)


type Server struct{
	URL string
	IsActive  bool
}


type LB struct{
Server *http.Server
Backends []Server
lock *sync.Mutex
}



var AvailableBackends=[]Server{
	{"http://127.0.0.1:3000",true},
	{"http://127.0.0.1:3003",true},
	{"http://127.0.0.1:3006",true},
}



//initialize the loadBalancer
func createLB(servers []Server,ctx context.Context)*LB{
lb:=&LB{lock: &sync.Mutex{}}
for _,server:=range servers{
	if server.IsActive{
		lb.Backends = append(lb.Backends, server)
	}
}
lb.Server=&http.Server{
	Addr: "127.0.0.1:8080",
  BaseContext: func(l net.Listener) context.Context {
		return ctx
	}, 
}
return lb
}



func (lb *LB) handler(w http.ResponseWriter, r *http.Request){
//used lock to maintain the order during parallel requests
lb.lock.Lock()
defer lb.lock.Unlock()
var server Server
server,lb.Backends=roundRobin(lb.Backends)
client:=&http.Client{}
req,err:=http.NewRequest(r.Method,server.URL,nil)
if err!=nil{
	fmt.Println(err.Error())
	http.Error(w,"LB:Error creating request",http.StatusInternalServerError)
 
}
req.Header=r.Header
resp,err:=client.Do(req)
if err!=nil{
	fmt.Println(err.Error())
  http.Error(w,"LB:Error forwarding request",resp.StatusCode)
 
}
var responseTxt string
for name,value:=range r.Header{
responseTxt+=fmt.Sprint("\n",name,":",value)
}
responseTxt+=fmt.Sprint("\n","Body:","\n")
body,err:=io.ReadAll(resp.Body)
if err!=nil{
	responseTxt+="Error Reading Response Body"
}
responseTxt+=string(body)
fmt.Fprint(w,responseTxt)
} 



func (lb *LB) healthCheck(interval int){
client:=&http.Client{}
for{
fmt.Println("starting periodic health check")	
for _,server:=range AvailableBackends{
req,err:= http.NewRequest("GET",fmt.Sprint(server.URL,"/healthCheck"),nil)
if err!=nil{
	fmt.Println("couldn't create request for server:",server.URL)
}
resp,err:=client.Do(req)
if err!=nil{
		fmt.Println("couldn't send health check request to server:",server.URL)
}else if !(resp.StatusCode>=200 && resp.StatusCode<=400){
server.IsActive=false
}else{
	fmt.Println("server:",server.URL,"passed health check!")
}

}
fmt.Println("periodic health check done")
	<-time.After(time.Minute*time.Duration(interval))
}
}




func roundRobin(servers []Server)(Server,[]Server){
selected:=servers[0]
servers=append(servers[1:],selected )
return  selected,servers
}



func main(){
var interval int
fmt.Println("starting LB server...")
fmt.Println("LB server running!")
fmt.Println("Enter the time period for health check(in minutes)")
fmt.Scanln(&interval)
ctx,cancel:=context.WithCancel(context.Background())
defer cancel()
lb:=createLB(AvailableBackends,ctx)
http.HandleFunc("/",lb.handler)
go lb.healthCheck(interval)
err:=lb.Server.ListenAndServe()
if errors.Is(err,http.ErrServerClosed){
	fmt.Println("LB server closed")
}else if err!=nil{
	fmt.Println(err.Error())
}

}