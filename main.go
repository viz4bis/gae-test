package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	//"errors"
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

type WeiboAccessToken struct {
  Token string
}

type DBValue struct {
  Weibo int64 
  Qq    int64 
  Stock float64 
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/setup", setup)
	http.HandleFunc("/fetch-and-store", fetch_and_store)
}

func root(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  q := datastore.NewQuery("DBValue").Limit(10)
  var values []DBValue
  _, err := q.GetAll(c, &values)
  if err != nil {
    c.Errorf("failed to get values : %v\n", err)
    return
  }
  fmt.Fprintf(w, "%v\n", values)
}

const setup_html = `
 <html>
 <body>
 <div>
 <form action="https://api.weibo.com/oauth2/authorize">
  <input type="hidden" name="client_id" value="2909741328">
  <input type="hidden" name="redirect_uri" value="http://ambient-depth-612.appspot.com/connect-weibo">
  <input type="hidden" name="response_type" value="code">
  <button>setup</button>
  </form></div>
 </body>
 </html>
`

func setup(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, setup_html)
}

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
    c.Errorf("%v\n", err)
    return
	}

	var kv map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&kv)
	if at, ok := kv["access_token"].(string); ok {
    entity := WeiboAccessToken {
      Token : at,
    }
    if key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "WeiboAccessToken", nil), &entity); err != nil {
      c.Errorf("failed to put : %v\n", err)
      return
    } else {
      c.Infof("encode : %v\n", key.Encode())
    }
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func lookup_weibo_access_token(r *http.Request) (string, error) {
  c := appengine.NewContext(r)
  if key, err := datastore.DecodeKey("ahNzfmFtYmllbnQtZGVwdGgtNjEych0LEhBXZWlib0FjY2Vzc1Rva2VuGICAgIC8oYIKDA"); err != nil {
    return "", err
  } else {
    var entity WeiboAccessToken
    if err := datastore.Get(c, key, &entity); err != nil {
      return "", err
    }
    return entity.Token, nil
  }
}

func fetch_and_store(w http.ResponseWriter, r *http.Request) {

  value := DBValue { Weibo : 0, Qq : 0, Stock : 0.0 }

  c := appengine.NewContext(r)
	client := urlfetch.Client(c)

  c.Infof("fetch-and-store")

  // weibo
  var err error
  var token string 
  if token, err = lookup_weibo_access_token(r); err != nil {
    c.Errorf("failed to get token : %v\n", err)
  }
	resp, err := client.Get("https://api.weibo.com/2/search/suggestions/users.json?q=uniqlo&access_token=" + token)
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

	resp, err = client.Get("https://api.weibo.com/2/users/show.json?uid=" + uid + "&access_token=" + token)
	if err != nil {
		fmt.Fprintf(w, "err : %s\n", err.Error())
		return
	}
	var kv map[string]interface{}
	dec = json.NewDecoder(resp.Body)
	dec.Decode(&kv)
	if v, ok := kv["followers_count"].(float64); ok {
		value.Weibo =  int64(v)
	}

  // qq by scraping
	resp, err = client.Get("http://e.t.qq.com/uniqlochina")
	if err != nil {
    c.Errorf("qq")
		return
	}

	tree, err := h5.New(resp.Body)
	if err != nil {
    c.Errorf("qq")
    return
	}

	t := transform.New(tree)
	t.Apply(func(node *html.Node) {
		for _, v := range node.Attr {
			if v.Val == "http://t.qq.com/uniqlochina/follower" {
				if value.Qq, err = strconv.ParseInt(node.FirstChild.Data, 10, 64); err != nil {
          c.Errorf("qq")
          return
				}
			}
		}
	}, "a.text_count.co_c_tx")

  // stock by yql
  resp, err = client.Get("https://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20yahoo.finance.quote%20where%20symbol%20in%20(%22FRCOY%22)&format=json&diagnostics=true&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys&callback=")
  if err != nil {
    c.Errorf("stock")
    return
  }
  var kv2 map[string]interface {}
  dec = json.NewDecoder(resp.Body)
  dec.Decode(&kv2)
  if kv3, ok := kv2["query"].(map [string] interface{}); ok {
    if kv4, ok := kv3["results"].(map [string] interface{}); ok {
      if kv5, ok := kv4["quote"].(map [string] interface{}); ok {
        if quote_s, ok := kv5["LastTradePriceOnly"].(string); ok {
          if value.Stock, err = strconv.ParseFloat(quote_s, 64); err != nil {
            c.Errorf("stock")
            return
          }
        } else {
          c.Errorf("stock")
          return
        }
      } else {
        c.Errorf("stock")
        return
      }
    } else {
      c.Errorf("stock")
      return
    }
  } else {
    c.Errorf("stock")
    return
  }

  if _, err := datastore.Put(c, datastore.NewKey(c, "DBValue", "", time.Now().Unix(), nil), &value); err != nil {
    c.Errorf("failed to put data : %v", err);
    return
  }
}
