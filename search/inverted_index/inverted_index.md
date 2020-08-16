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

#### 倒排索引的加速
在倒排索引的检索过程中，两个posting list求交是最重要，最耗时的操作。有以下方法可以加速:

##### 跳表法加速倒排索引
假设posting list A 中的元素为 <1,2,3,4,5,6,7,8……，1000>，这 1000 个元素是按照从 1 到 1000 的顺序递增的。而 posting list B 中的元素，只有 <1,500,1000>3 个。那么按照我们之前的归并方法，在找到元素1以后，还需要再遍历 498 次链表，才能找到第二个相同元素 500。

![post1](https://github.com/checkking/notes/blob/master/imgs/pos_1.png)

我们可以通过将有序链表改成跳表，这样，在posting list A中，我们从第2个元素遍历到第500个元素，只需要 log(498) 次的量级，会比链表快得多。


![post2](https://github.com/checkking/notes/blob/master/imgs/post2.png)

其实可以相互二分查找的，posting list A中，拿500在posting list B中二分查找，posting list B中，拿1000在posting list A中二分查找. 在实际的系统中，如果posting list可以都存储在内存中，并且变化不频繁的话，可以用*可变数组*来代替链表。 这样，对于两个posting list求交集，我们同样可以使用相互二分查找，进行归并，并且可以利用CPU的局部性提高性能。

##### 哈希表法加速倒排索引

假如posting list A的元素比posting list B的元素多很多，我们可以提前将posting list A存入哈希表中. 这样B中的每个元素查找A就是O(1)， 如果B有m个元素，就是O(m)。在真实场景中，提前把那些频繁需要求交的posting list用hash表存储，当然，为了这些posting list也能够有遍历能力，posting list本省也是要保留的。

##### 位图法加速倒排索引

posting list 用位图来存储, 但是有以下局限:
1. 位图法仅适用于只存储 ID 的简单的 posting list。如果 posting list 中需要存储复杂的对象，就不适合用位图来表示 posting list 了。
2. 位图法仅适用于 posting list 中元素稠密的场景。对于 posting list 中元素稀疏的场景，使用位图的运算和存储开销反而会比使用链表更大。
3. 位图法会占用大量的空间。尽管位图仅用 1 个 bit 就能表示元素是否存在，但每个 posting list 都需要表示完整的对象空间。如果 ID 范围是用 int32 类型的数组表示的，那一个位图的大小就约为 512M 字节。如果我们有 1 万个 key，每个 key 都存一个这样的位图，那就需要 5120G 的空间了。

##### 升级版位图：Roaring Bitmap

Roaring Bitmap将一个32位的整数分为两个部分，一部分是高16位，另一部分是低16位。对于高16位，Roaring Bitmap将它存储在一个有序数组中，有序数组的每个值都是一个“桶”; 对于低16位，将它存储在一个2^16的位图中，相应位置置为1。 这样每个通都会对应一个2^16的位图, 也就是8K。

![roaring-map](https://github.com/checkking/notes/blob/master/imgs/roaring_map.png)

要判断一个元素是否在Roaring Bitmap中，需要有两步，先在用高16位在有序数组中通过二分查找法,找到对应的桶， 然后用低16位在桶对应的bitmap找到对应的位置.第一步查找由于是数组二分查找，因此时间代价是 O（log n）；第二步是位图查找，因此时间代价是 O(1)。 其实，bitmap在数据很小的时候，可以退化成有序数组，比如元素小于4096，我们可以用有序数组来代替bitmap节省存储空间。


![roaring-map2](https://github.com/checkking/notes/blob/master/imgs/roaring_map2.png)
