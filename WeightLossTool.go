package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

var(
	location="beijing"
	key="******************"
)
type AirQuality struct {
	HeWeather6 []struct {
		Basic struct {
			Cid        string `json:"cid"`
			Location   string `json:"location"`
			ParentCity string `json:"parent_city"`
			AdminArea  string `json:"admin_area"`
			Cnty       string `json:"cnty"`
			Lat        string `json:"lat"`
			Lon        string `json:"lon"`
			Tz         string `json:"tz"`
		} `json:"basic"`
		Update struct {
			Loc string `json:"loc"`
			Utc string `json:"utc"`
		} `json:"update"`
		Status     string `json:"status"`
		AirNowCity struct {
			Aqi     string `json:"aqi"`
			Qlty    string `json:"qlty"`
			Main    string `json:"main"`
			Pm25    string `json:"pm25"`
			Pm10    string `json:"pm10"`
			No2     string `json:"no2"`
			So2     string `json:"so2"`
			Co      string `json:"co"`
			O3      string `json:"o3"`
			PubTime string `json:"pub_time"`
		} `json:"air_now_city"`
		AirNowStation []AirNowStation `json:"air_now_station"`
	} `json:"HeWeather6"`
}
type AirNowStation struct {
	AirSta  string `json:"air_sta"`
	Aqi     string `json:"aqi"`
	Asid    string `json:"asid"`
	Co      string `json:"co"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Main    string `json:"main"`
	No2     string `json:"no2"`
	O3      string `json:"o3"`
	Pm10    string `json:"pm10"`
	Pm25    string `json:"pm25"`
	PubTime string `json:"pub_time"`
	Qlty    string `json:"qlty"`
	So2     string `json:"so2"`
}
type Weather struct {
	HeWeather6 []struct {
		Basic struct {
			Cid        string `json:"cid"`
			Location   string `json:"location"`
			ParentCity string `json:"parent_city"`
			AdminArea  string `json:"admin_area"`
			Cnty       string `json:"cnty"`
			Lat        string `json:"lat"`
			Lon        string `json:"lon"`
			Tz         string `json:"tz"`
		} `json:"basic"`
		Now Now `json:"now"`
		Status string `json:"status"`
		Update struct {
			Loc string `json:"loc"`
			Utc string `json:"utc"`
		} `json:"update"`
	} `json:"HeWeather6"`
}
type Now struct {
	CondCode string `json:"cond_code"`
	CondTxt  string `json:"cond_txt"`
	Fl       string `json:"fl"`
	Hum      string `json:"hum"`
	Pcpn     string `json:"pcpn"`
	Pres     string `json:"pres"`
	Tmp      string `json:"tmp"`
	Vis      string `json:"vis"`
	WindDeg  string `json:"wind_deg"`
	WindDir  string `json:"wind_dir"`
	WindSc   string `json:"wind_sc"`
	WindSpd  string `json:"wind_spd"`
}
func doSendEmail(user,password,host,to,subject,body,content_type string)error{
	server:=strings.Split(host,":")
	auth:=smtp.PlainAuth("",user,password,server[0])
	message := []byte("To:" + to +"\r\nFrom: " + user + "<"+
		user + ">\r\nSubject: "+ subject + "\r\n" +
		content_type + "\r\n\r\n" + body)
	sendTo := strings.Split(to,";")
	err:=smtp.SendMail(host,auth,user,sendTo,message)

	return err
}
func sendEmail(body string){
	from:="18653059888@163.com"
	password:="********"
	host:="smtp.163.com:25"
	to:="1035780360@qq.com"
	contentType:="Content-Type: text/plain; charset=UTF-8"
	subject:="跑起来吧"
	err:=doSendEmail(from, password, host, to,subject, body,contentType)
	if err!=nil{
		fmt.Println(err)
	}
}
/*
	处理Get类型的请求，返回响应数据
 */
func doHttpGetRequest(url string)(result string, err error){
	resp,err:=http.Get(url)
	if err!=nil{
		return "", err
	}
	defer resp.Body.Close()
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		return "", err
	}
	return string(body), err
}
/*
	获取当前空气质量
 */
func getAirQualityNow()([]interface{},error){
	url:="https://free-api.heweather.net/s6/air/now?"
	result,err:=doHttpGetRequest(url+"location="+location+"&key="+key)
	if err!=nil{
		return nil, err
	}
	var data AirQuality
	err1:=json.Unmarshal([]byte(result), &data)
	if err1!=nil{
		return nil,err1
	}
	var dataSlice=make([]interface{},0,15)
	for _,item:= range data.HeWeather6{
		for _,item1:=range item.AirNowStation{
			dataSlice= append(dataSlice, item1)
		}
	}
	return dataSlice,err
}
/*
	从获取到的空气质量数据中过滤出需要的信息
 */
func filterAirQualityInformation(stations []interface{},information *[]interface{},wg *sync.WaitGroup){
	defer wg.Done()
	data,err:=getAirQualityNow()
	if err!=nil{
		return
	}
	for _,item:=range data{
		for _,item1:=range stations {
			if item.(AirNowStation).AirSta == item1.(string) &&(item.(AirNowStation).Qlty=="优"||item.(AirNowStation).Qlty=="良"){
				*information=append(*information,item)
			}
		}
	}
}
/*
	获取当前天气状况
 */
func getWeatherNow(information *[]interface{},wg *sync.WaitGroup)(){
	defer wg.Done()
	url:="https://free-api.heweather.net/s6/weather/now?"
	result,err:=doHttpGetRequest(url+"location="+location+"&key="+key)
	if err!=nil{
		return
	}
	var data Weather
	err1:=json.Unmarshal([]byte(result), &data)
	if err1!=nil{
		return
	}
	for _,item:=range data.HeWeather6{
		*information=append(*information,item.Now)
	}
}
func startWork(){
	var wg sync.WaitGroup
	wg.Add(2)
	airQualityInformation:=make([]interface{},0,10)
	weatherInformation:=make([]interface{},0,5)
	Slice:=make([]interface{},0,10)
	Slice=append(Slice,"海淀区万柳")
	go filterAirQualityInformation(Slice,&airQualityInformation,&wg)
	go getWeatherNow(&weatherInformation,&wg)
	wg.Wait()
	if len(airQualityInformation)!=0 && (weatherInformation[0].(Now).CondTxt=="多云"||weatherInformation[0].(Now).CondTxt=="晴"){
		content:="今天空气质量  "+ airQualityInformation[0].(AirNowStation).Qlty+" ,天气  "+weatherInformation[0].(Now).CondTxt +" ,赶紧起床去跑步！ "
		sendEmail(content)
	}
}
/*
	开启计时器
 */
func startTimer(f func()){
	go func() {
		for   {
			f()
			currentTime:=time.Now()
			//计算下一个6:30
			next:=currentTime.Add(time.Hour*24)
			next=time.Date(next.Year(),next.Month(),next.Day(),6,30,0,0,next.Location())
			t:=time.NewTimer(next.Sub(currentTime))
			<-t.C
		}
	}()
}
func main(){
	startTimer(startWork)
	exit:=make(chan int)
	<-exit
}