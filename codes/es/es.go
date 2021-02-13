package comm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/golang/glog"
	"time"
)

var ESClient *elasticsearch.Client

func InitESClient(addr, user, pwd string) error {
	cfg := elasticsearch.Config{
		Addresses:            []string{
			fmt.Sprintf("http://%s", addr),
		},
		Username:              user,
		Password:              pwd,
	}
	var err error
	ESClient, err = elasticsearch.NewClient(cfg)
	return err
}

func QueryES(dsl, esIndex, docType string) (map[string]interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			panicStr := commUtils.GetStackInfo()
			glog.Errorf("ES Query failed, err:%v panic:%s", err, panicStr)
		}
	}()

	var buf *bytes.Buffer
	buf = bytes.NewBuffer([]byte(dsl))

	esRest :=  map[string]interface{}{}

	res, err := ESClient.Search(
		ESClient.Search.WithContext(context.Background()),
		ESClient.Search.WithIndex(esIndex),
		ESClient.Search.WithDocumentType(docType),
		ESClient.Search.WithBody(buf),
		ESClient.Search.WithTrackTotalHits(true),
		ESClient.Search.WithTimeout(time.Millisecond * 800),
	)
	if err != nil {
		glog.Errorf("failed to search:%s", err.Error())
		return esRest, err
	}

	defer res.Body.Close()
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			glog.Errorf("Error parsing the response body: %s", err.Error())
			return esRest, err
		} else {
			// Print the response status and error information.
			glog.Errorf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return esRest, errors.New(fmt.Sprintf("es query failed, err:%s", e["error"].(map[string]interface{})["reason"]))
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&esRest); err != nil {
		glog.Errorf("Error parsing the response body: %s", err)
		return esRest, err
	}
	return esRest, nil
}

// 改造过滤条件
func excludeMaskTermList(channelId int, excludeItems map[string][]string) []map[string]interface{} {
	mustNotList := make([]map[string]interface{},0)
	// mask in must_not
	mustNotList = append(mustNotList, stringTerm("mask_channel", strconv.Itoa(channelId)))

	// exclude docs in must_not
	if excludeDocs, ok := excludeItems["docid"]; ok {
		excludeDocs = commUtils.FilterNullStr(excludeDocs)
		if len(excludeDocs) > 0 {
			mustNotList = append(mustNotList, excludeTerm(excludeDocs))
		}
	}
	// exclude disease in must_not
	if excludeDiseases, ok := excludeItems["disease"]; ok {
		excludeDiseases = commUtils.FilterNullStr(excludeDiseases)
		for _, dName := range excludeDiseases {
			mustNotList = append(mustNotList, stringTerm("disease", dName))
		}
	}
	// exclude category_two in must_not
	if excludeClass2, ok := excludeItems["class2"]; ok {
		excludeClass2 = commUtils.FilterNullStr(excludeClass2)
		for _, cName := range excludeClass2 {
			cateTags := strings.Split(cName, "+")
			if len(cateTags) == 2 {
				cateMustList := []map[string]interface{}{stringTerm("category_one", cateTags[0]),
						stringTerm("category_two", cateTags[1])}
				mustNotList = append(mustNotList, boolQueryTerm("must", cateMustList))
			}
		}
	}

	// exclude video，keep docs with default value, so in must_not
	if channelId == 200 {
		mustNotList = append(mustNotList, intTerm("video_direction", 2))  // 竖屏视频过滤
		mustNotList = append(mustNotList, rangeTerm("video_duration", "lte", 6))  // 少于6秒视频过滤
	}
	return mustNotList
}

// 查询DSL, bool查询
func queryDsl(mustList, shouldList, mustNotList, filterList []map[string]interface{}) map[string]interface{} {
	boolPart := map[string]interface{}{"must":mustList,"should":shouldList,"must_not":mustNotList,"filter":filterList}
	queryPart := map[string]interface{}{"bool":boolPart}
	return queryPart
}

// 构造dsl, 包含过滤条件
func DslNewestRecall(channelId int, docNum int, excludeItems map[string][]string) (string, error) {
	/*	channelId: switch diff channel, also need to exclude the masked docs;
		docNum: recall number;
		excludeItems: empty means no need to exclude, {"docid":[], "disease":[], "class2":[]};
	*/
	mustList := make([]map[string]interface{},0)
	shouldList := make([]map[string]interface{},0)

	// must not term list
	mustNotList := excludeMaskTermList(channelId, excludeItems)
	// diff doc source for diff channels
	filterList := docFilterTermList(channelId, true)

	// query dsl
	queryPart := queryDsl(mustList, shouldList, mustNotList, filterList)

	// dsl string
	dslStr, err := queryDslString(queryPart, channelId, docNum, "newest")
	return dslStr, err
}

// 聚合查询, json
POST xx_index/article_info/_search
{
  "query": {
    "bool": {
      "filter":[
      ],
      "must": [
      {
        "bool":{
          "should":[
            {"term":{"disease":"痛风"}},
            {"term":{"disease":"高血压"}},
            {"term":{"disease":"糖尿病"}}
          ]
        }
      }
    ]
    }
  },
  "size":0,
  "aggs": {
    "tag_agg": {
      "filters": {
        "filters": {
          "1": {"bool": {"should": [{"term":{"disease":"痛风"}},
                  {"term":{"disease":"高血压"}}]}},
          "2": {"term":{"disease":"糖尿病"}},
          "3": {"bool": {"should": []}},
          "4": {"term":{"disease":"糖尿病"}}
        }
      },
      "aggs":{
        "src_agg": {
          "filters": {
            "filters": {
              "video": {"term":{"ctype":8}},
              "qa": {"term":{"ctype":11}}
            }
          },
          "aggs":{
            "tag_agg_hits":{
              "top_hits":{
                "size": 2,
                "_source": ["docid","disease","createtime", "ctype"],
                "sort": [{"_score":{"order":"desc"}},
                         {"createtime":{"order":"desc"}}],
                "track_scores": true
              }
            }
          }
        }
      }
    }
  }
}
