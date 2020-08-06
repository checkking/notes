### 布隆过滤器
#### 用位图来方便查找是否存在
假设要查找一个用户ID是否已经注册过, 或者其他的查询是否存在的需求，用位图空间换时间的方式一种常见的方案。
将要查找的内容转成一个整数n，通过查找bitmap的第n位是否为1排判断是否存在。

![bitmap](https://github.com/checkking/notes/blob/master/imgs/bitmap.png)

假设用char数组来实现bitmap, 一个char占8个bit, 判断n是否在bitmap中，可以先找到在数据的第几个位置和具体的第几个位。具体实现见golang代码.

#### 用布隆过滤器来节省存储空间
布隆过滤器最大的特点，就是对一个对象使用多个哈希函数。如果我们使用了 k 个哈希函数，就会得到 k 个哈希值，也就是 k 个下标，我们会把数组中对应下标位置的值都置为 1。布隆过滤器和位图最大的区别就在于，我们不再使用一位来表示一个对象，而是使用 k 位来表示一个对象。这样两个对象的 k 位都相同的概率就会大大降低，从而能够解决哈希冲突的问题了。

![bloomfilter](https://github.com/checkking/notes/blob/master/imgs/bloomfilter.png)

bloomfilter存在假正的问题(False Positive)， 也就是不存在一定是真的，判断存在不一定是真的。这个错误率是受位图大小、元素个数和hash函数个数影响的。

计算最优哈希函数个数的数学公式: 哈希函数个数 k = (m/n) * ln(2)。其中 m 为 bit 数组长度，n 为要存入的对象的个数。实际上，如果哈希函数个数为 1，且数组长度足够，布隆过滤器就可以退化成一个位图。

#### 用布隆过滤器做曝光去重
对于一些业务场景，比如推荐，推荐系统需要将用户看过的文章做过滤处理，不再推荐, 这就需要记录用户已经看过的信息id. 一个比较简单的方案是给每个用户存储一个明文内容曝光id列表，结合缓存进行存储，比如存储在redis中。

![uidlist](https://github.com/checkking/notes/blob/master/imgs/uidlist.png?raw=true)

这种方案易于实现，缺点就是存储空间大，比如一个文章的id有14个字节，如果限定记录条数为5000，每个用户大概需要70K左右的存储空间，100G也就够存储100多万个用户记录。 并且比较效率很低，需要遍历用户的list才能判断是否曝光过. 用布隆过滤器的方案可以节省内存，但是需要结合特定业务场景。

一般业务场景下，对于用户曝光记录，不太可能无线增加，可以为每个用户保存有限的曝光记录就可以, 并且一般内容具有时效性，用户的曝光记录可以在一定冷却时间之后进行淘汰。我们可以给每个用户限定5000条曝光记录，并且用户如果5天内没有新的曝光，则将他的曝光记录淘汰掉。另外，如果一开始就为每个用户分配一个能容纳5000的布隆过滤器，这样也有点浪费，可以采用分片的方式，一开始只为用户分配容纳1000的布隆过滤器，后面慢慢增至5个布隆过滤器，各布隆过滤器之间以列表的形式保存。


具体代码:

```
const (
   bloomfilterNum = 5          //每个人允许布隆过滤器最大个数
   maxKeyNum = 1000            //单个布隆最多能存元素个数
   FalsePositives = 0.001      //误识别率
)

//设置曝光记录
func SetExposed(uid string, ids []string) (error)  {
   if len(uid) < 2 || len(ids) == 0 {
      return errors.New("params error")
   }
   //预估布隆数据块大小和映射函数个数
   numBits, numHashFunc := bloomfilter.EstimateParameters(maxKeyNum, FalsePositives)
   //用户曝光记录key
   key := uid
   //当前已经设置文章id数量
   num := 0
   //判断是否需要新增
   isNew := true
   //初始化一个布隆过滤器
   bf := bloomfilter.New(numBits, numHashFunc)
   rds := redis.New("local")
   exposedData, err := redis.String(rds.Do(nil, "LINDEX", key, 0))
   //如果已经有记录则先加载
   if err == nil && exposedData != ""{
      arr := strings.Split(exposedData, "::")
      if len(arr) == 2 {
         num,err = strconv.Atoi(arr[0])
         if err == nil && num+len(ids) < maxKeyNum+10{
            decoded, err := base64.StdEncoding.DecodeString(arr[1])
            if err == nil{
			  //加载已有曝光记录
               bf = bloomfilter.NewFromBytes(decoded, numHashFunc)
            }
            isNew = false
         }else{
            num = 0
         }
      }
   }

   //添加新的曝光记录
   for _,id := range ids{
      if !bf.Test([]byte(id)){
         bf.Add([]byte(id))
         num++
      }
   }
   //开头代表此块已用容量，格式：数量::bloomfilter的字符串
   encoded := fmt.Sprintf("%d::%s", num, base64.StdEncoding.EncodeToString(bf.ToBytes()))
   if encoded != exposedData{
      if isNew == true{
         rds.Do(nil, "LPUSH", key, encoded)
      }else{
         rds.Do(nil, "LSET", key, 0, encoded)
      }
      rds.Do(nil, "LTRIM", key, 0, bloomfilterNum-1)
      rds.Do(nil, "EXPIRE", key, 3600*24*bloomfilterNum)
   }
   return err
}

//获取曝光记录，返回布隆过滤器列表
func GetExposed(uid string) (bfs []*bloomfilter.BloomFilter, err error) {
   if len(uid) < 2 {
      return nil, errors.New("params error")
   }
   _, numHashFunc := bloomfilter.EstimateParameters(maxKeyNum, FalsePositives)
   //用户曝光记录key
   key := uid
   rds := redis.New("local")
   exposedData, err := redis.Strings(rds.Do(nil, "LRANGE", key, 0, bloomfilterNum-1))
   if err == nil{
      bfs = make([]*bloomfilter.BloomFilter, 0)
      for _,item := range exposedData {
         arr := strings.Split(item, "::")
         if len(arr) == 2 && arr[1] != ""{
            decoded, err := base64.StdEncoding.DecodeString(arr[1])
            if err == nil {
               bf := bloomfilter.NewFromBytes(decoded, numHashFunc)
               bfs = append(bfs, bf)
            }
         }
      }
      return bfs, err
   }

   return nil, err
}

// Exists 判断key是否已经曝光过
func Exists(bfs []*bloomfilter.BloomFilter, key string) bool {
	for _, bf := range bfs {
		if bf.Test([]byte(key)) {
			return true
		}
	}
	return false
}
```
