package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Backend struct {
	Server *http.Server
}



func createBE(ctx context.Context,addr string,mux  *http.ServeMux,name string)*Backend{
BE:=&Backend{}
BE.Server=&http.Server{
Addr: addr,
Handler: mux,
BaseContext: func(l net.Listener) context.Context {
	newCtx:=context.WithValue(ctx,"name",name)
	return newCtx
},
}
return BE
}


func healthCheckHandler(w http.ResponseWriter, r *http.Request){
w.WriteHeader(http.StatusOK)
}


func  handler(w http.ResponseWriter, r *http.Request){
respch:=make(chan string)
ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second)
defer cancel()
go returnResponse(ctx,respch,r)
select{
case <-ctx.Done():
  fmt.Fprint(w,fmt.Sprintln(ctx.Err().Error()))
case result:=<-respch:
	fmt.Fprint(w,result)
} 

}


func returnResponse(ctx context.Context,ch chan string,r *http.Request){
name:=ctx.Value("name").(string)
//test server response 
ch<-fmt.Sprint("\n","server name:",name,"\n","Received request from:",r.RemoteAddr,"\n",r.Method,r.URL.Path,r.Proto,
"\n","Host:",r.Host,"\n","user-agent:",r.UserAgent(),"\n","Accept:",r.Header.Get("Accept"),"\n","Replied with a hello message")
}

func main(){
fmt.Println("Starting the servers")
mux:=http.NewServeMux()
ctx,cancel:=context.WithCancel(context.Background())
server1:=createBE(ctx,"127.0.0.1:3000",mux,"server1")
server2:=createBE(ctx,"127.0.0.1:3003",mux,"server2")
server3:=createBE(ctx,"127.0.0.1:3006",mux,"server3")
mux.HandleFunc("/",handler)
mux.HandleFunc("/healthCheck",healthCheckHandler)
go func(cancel context.CancelFunc){
	fmt.Println("running server 1")
	defer cancel()
	err:=server1.Server.ListenAndServe()
	if errors.Is(err,http.ErrServerClosed){
    fmt.Print("BE server closed")
	}else if err!=nil{
		fmt.Print(err.Error())
	}
}(cancel)


go func(cancel context.CancelFunc){
		fmt.Println("running server 2")
	defer cancel()
	err:=server2.Server.ListenAndServe()
	if errors.Is(err,http.ErrServerClosed){
    fmt.Print("BE server closed")
	}else if err!=nil{
		fmt.Print(err.Error())
	}
}(cancel)


go func(cancel context.CancelFunc){
		fmt.Println("running server 3")
	defer cancel()
	err:=server3.Server.ListenAndServe()
	if errors.Is(err,http.ErrServerClosed){
    fmt.Print("BE server closed")
	}else if err!=nil{
		fmt.Print(err.Error())
	}
}(cancel)
<-ctx.Done()
}