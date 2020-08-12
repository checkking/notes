### 倒排索引

根据具体内容名或者属性来检索文章的结构，称作*倒排索引(inverted index)*， 在倒排索引中，key的集合叫做字典(Dictionary)， 一个key后面对应的记录集合叫做记录列表(Posting list).

![interved-index](https://github.com/checkking/notes/blob/master/imgs/inverted_idnex.png)

#### 如何创建倒排索引
1. 给每个文档编号， 作为唯一的标识，并且排好序, 然后遍历文档.
2. 解析当前文档中的每个关键字，生成<关键字, 文档ID, 关键字位置>这样的数据对。
3. 将关键字作为key插入哈希表。如果key已经在哈希表中，则在posting list后面追加节点，记录该文档ID(关键字的位置信息如果需要，也可以一并记录在节点中); 如果哈希表中没有这个key, 就直接插入key, 并创建posting list和对应节点。
4. 重复2-3步，处理完所有文档，完成倒排索引的创建.

![create-inverted-index](https://github.com/checkking/notes/blob/master/imgs/create_interved_index.png)

#### 检索同时包含关键字"hello", "world"两个key的文档
查询"hello", "world"两篇文档，以这两个词作为key检索倒排索引查找，得到两个posting list，我们需要同时包含两个关键字的posting list，就需要将这两个posting list的相同元素找出来。 如果posting list列表不是顺序的，查找操作的复杂度就是O(m\*n)，
但是，如果posting list的元素是有序的，则查找和用归并排序，复杂度就降为O(m+n)。

第一步, 使用指针p1和p2分别指向有序列表A和B的第一个元素。

第二步， 对比p1和p2所指元素是否相同:

* 两者id相同，说明是公共元素，直接将该节点加入归并结果。 然后p1++, p2++;

* p1的id < p2的id, p1++

* p1的id > p2的id, p2++

重复第2步，直到p1或p2移到链表结尾为止.


对于查找或的关系, 或者多个key查询的场景，也可以采用类似的方法。


说明： 这个只是一个简易的方法，对于数据量大的情况，或者真实的工业界场景，很显然是有性能问题的。


