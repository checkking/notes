### 倒排索引

根据具体内容名或者属性来检索文章的结构，称作*倒排索引(inverted index)*， 在倒排索引中，key的集合叫做字典(Dictionary)， 一个key后面对应的记录集合叫做记录列表(Posting list).

![interved-index](https://github.com/checkking/notes/blob/master/imgs/inverted_idnex.png)

#### 如何创建倒排索引
1. 给每个文档编号， 作为唯一的标识，并且排好序, 然后遍历文档.
2. 解析当前文档中的每个关键字，生成<关键字, 文档ID, 关键字位置>这样的数据对。
3. 将关键字作为key插入哈希表。如果key已经在哈希表中，则在posting list后面追加节点，记录该文档ID(关键字的位置信息如果需要，也可以一并记录在节点中); 如果哈希表中没有这个key, 就直接插入key, 并创建posting list和对应节点。
4. 重复2-3步，处理完所有文档，完成倒排索引的创建.

![create-inverted-index](https://github.com/checkking/notes/blob/master/imgs/create_interved_index.png)
