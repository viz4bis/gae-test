package main

import (
	"appengine"
	"appengine/urlfetch"
  //"database/sql"
	"time"
	"encoding/json"
	"fmt"
	//"io"
	//"io/ioutil"
	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go-html-transform/html/transform"
	"code.google.com/p/go.net/html"
	"net/http"
	"net/url"
	"strconv"
  //_ "github.com/iitaku/go-yql"
	//"code.google.com/p/go-html-transform/css/selector"
	//"html/template"
	//"github.com/PuerkitoBio/goquery"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/connect-weibo", connect_weibo)
	http.HandleFunc("/view", view)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, root_html)
}

const root_html = `
 <html>
 <body>
 <div>
 <form action="https://api.weibo.com/oauth2/authorize">
  <input type="hidden" name="client_id" value="2909741328">
  <input type="hidden" name="redirect_uri" value="http://ambient-depth-612.appspot.com/connect-weibo">
  <input type="hidden" name="response_type" value="code">
  <button>View</button>
  </form></div>
 </body>
 </html>
`

func connect_weibo(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	resp, err := client.PostForm(
		"https://api.weibo.com/oauth2/access_token",
		url.Values{"client_id": {"2909741328"},
			"client_secret": {"61d75f2cceb35c8d95dd7185afb1dd5c"},
			"grant_type":    {"authorization_code"},
			"code":          {r.URL.Query()["code"][0]},
			"redirect_uri":  {"http://ambient-depth-612.appspot.com/connect-weibo"}})

	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}

	var kv map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&kv)
	if at, ok := kv["access_token"].(string); ok {
		cookie := http.Cookie{Name: "access_token", Value: at}
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, "/view", http.StatusMovedPermanently)
}

func view(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

  // weibo
	at, err := r.Cookie("access_token")
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}

	resp, err := client.Get("https://api.weibo.com/2/search/suggestions/users.json?q=uniqlo&access_token=" + at.Value)
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}
	var arr []interface{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&arr)
	var uid string
	if len(arr) != 0 {
		if kv, ok := arr[0].(map[string]interface{}); ok {
			if v, ok := kv["uid"].(float64); ok {
				uid = strconv.FormatFloat(v, 'f', 0, 64)
			}
		} else {
			fmt.Fprintf(w, "err : %s\n", err.Error())
			return
		}
	}

	resp, err = client.Get("https://api.weibo.com/2/users/show.json?uid=" + uid + "&access_token=" + at.Value)
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}
	var kv map[string]interface{}
	dec = json.NewDecoder(resp.Body)
	dec.Decode(&kv)
	if v, ok := kv["followers_count"].(float64); ok {
		fmt.Fprintf(w, "weibo follower : %v\n", uint64(v))
	}

  // qq by scraping
	resp, err = client.Get("http://e.t.qq.com/uniqlochina")
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}

	tree, err := h5.New(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}

	t := transform.New(tree)
	var qq_follower uint64
	t.Apply(func(node *html.Node) {
		for _, v := range node.Attr {
			if v.Val == "http://t.qq.com/uniqlochina/follower" {
				if qq_follower, err = strconv.ParseUint(node.FirstChild.Data, 10, 64); err != nil {
					qq_follower = 0
				}
			}
		}
	}, "a.text_count.co_c_tx")
	fmt.Fprintf(w, "qq follower : %v\n", qq_follower)

  // stock by yql
  resp, err = client.Get("https://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20yahoo.finance.quote%20where%20symbol%20in%20(%22FRCOY%22)&format=json&diagnostics=true&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys&callback=")
  if err != nil {
    fmt.Fprintln(w, "%v\n", err)
    return
  }
  var kv2 map[string]interface {}
  dec = json.NewDecoder(resp.Body)
  dec.Decode(&kv2)
  var quote float64
  if kv3, ok := kv2["query"].(map [string] interface{}); ok {
    if kv4, ok := kv3["results"].(map [string] interface{}); ok {
      if kv5, ok := kv4["quote"].(map [string] interface{}); ok {
        if quote_s, ok := kv5["LastTradePriceOnly"].(string); ok {
          if quote, err = strconv.ParseFloat(quote_s, 64); err != nil {
            quote = 0
          }
        }
      }
    }
  }
  fmt.Fprintf(w, "quote : %v\n", quote)
}

func cron(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
  c.Infof("%v\n", time.Now())
}
