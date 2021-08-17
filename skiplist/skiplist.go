/*****************************
// Time : 2021/4/19 14:34
// Author : shitao
// File : skiplist
// Software: GoLand
// Description: 跳表核心实现
*****************************/

package skiplist

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const SkiplistMaxlevel = 32 // 最高层数限制
const SkiplistP = 2         // 随机数对2取模,奇偶各占一半 相当1/2概率加一层

// Node 跳表节点
type Node struct {
	NodeHeader

	value interface{}
	key   interface{}
	score float64 // 统一比较值,由key转化得来

	prev         *Node     // 前向指针
	prevTopLevel *Node     // 指向前置节点顶层
	list         *Skiplist // 该节点所处跳表实例
}
type NodeHeader struct {
	levels []*Node // 保存各层指向下一节点的指针 长度为当前节点层数
}
type Skiplist struct {
	NodeHeader

	comparable Comparable
	rand       *rand.Rand

	maxLevel int
	length   int
	back     *Node
}

func SkiplistInit(key Comparable) *Skiplist {
	list := Skiplist{
		NodeHeader: NodeHeader{
			levels: make([]*Node, SkiplistMaxlevel),
		},
		comparable: key,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		maxLevel:   SkiplistMaxlevel,
	}
	list.back = nil
	list.length = 0
	list.levels = make([]*Node, len(list.levels))
	return &list
}

// 新建节点
func newNode(list *Skiplist, level int, score float64, key, value interface{}) *Node {
	Node := Node{
		NodeHeader: NodeHeader{
			levels: make([]*Node, level),
		},
		value: value,
		key:   key,
		score: score,
		list:  list,
	}
	return &Node
}

// Next returns next node
func (node *Node) Next() *Node {
	if len(node.levels) == 0 {
		return nil
	}

	return node.levels[0]
}

// Prev 返回前置节点
func (node *Node) Prev() *Node {
	return node.prev
}

// NextLevel 返回指定层的下一节点
func (node *Node) NextLevel(level int) *Node {
	if level < 0 || level >= len(node.levels) {
		return nil
	}

	return node.levels[level]
}

// PrevLevel 返回指定层的前一节点
func (node *Node) PrevLevel(level int) *Node {
	if level < 0 || level >= len(node.levels) {
		return nil
	}

	if level == 0 {
		return node.prev
	}

	if level == len(node.levels)-1 {
		return node.prevTopLevel
	}

	prev := node.prev

	for prev != nil {
		if level < len(prev.levels) {
			return prev
		}

		prev = prev.prevTopLevel
	}

	return prev
}

func (list *Skiplist) Empty() bool {
	return list.length == 0
}

// Key 隐藏key, 避免对key的误修改
func (node *Node) Key() interface{} {
	return node.key
}
func (node *Node) Value() interface{} {
	return node.value
}

// Score 同样隐藏
func (node *Node) Score() float64 {
	return node.score
}

// Level 返回跳表实例最高层
func (node *Node) Level() int {
	return len(node.levels)
}

// Element 返回
func (header *NodeHeader) Element() *Node {
	return (*Node)(unsafe.Pointer(header))
}

func (list *Skiplist) Front() (front *Node) {
	return list.levels[0]
}

func (list *Skiplist) Back() (back *Node) {
	return list.back
}

// Print 跳表数据结构打印
func (list *Skiplist) Print() {
	fmt.Printf("\n*****Skip List*****\n")
	for i := 0; i <= list.maxLevel; i++ {
		node := list.NodeHeader.levels[i]
		fmt.Printf("Level : %d", i)
		for node != nil {
			fmt.Println(node.Key(), ":", node.Value(), ";")
			node = node.levels[i]
		}
		fmt.Println()
	}
}

// Insert k-value
func (list *Skiplist) Insert(key, value interface{}) (node *Node, err error) {
	score := list.calcScore(key)   // key转为统一标识值
	level := list.getRandomLevel() // 获取该节点层数
	node = newNode(list, level, score, key, value)

	if list.Empty() {
		// 直接插入
		for i := 0; i < level; i++ {
			// 对每层添加指向下一元素的指针
			list.levels[i] = node
		}
		list.back = node
		list.length++
		return
	}
	// 找到第一个小于key的节点
	curMaxLevel := len(list.levels) // 跳表当前最高层,非理论最高层(这里就是最高层,预先分配内存)
	var allocMem [SkiplistMaxlevel]*NodeHeader
	var prevHead []*NodeHeader
	prevHead = allocMem[:curMaxLevel] // prevHead[i]保存的是每层的跳转点
	curHead := &list.NodeHeader       // dummy头
	for i := curMaxLevel - 1; i >= 0; {
		// 当前层无元素或大于要插入的key则转入下一层寻找
		prevHead[i] = curHead
		// 在当前层中比较，直到查找元素小于key
		for next := curHead.levels[i]; next != nil; next = curHead.levels[i] {
			if comp := list.compare(score, key, next); comp <= 0 {
				// key已存在或者字符串哈希冲突
				// 更新后直接返回
				if comp == 0 {
					// 冲突,更新值
					node = next
					node.value = value
					// 暂且不算错误
					return node, nil
				}
				// 下一节点key大于待插入节点的key
				// 既将newNode插入next之前
				break
			}

			// 跳到当前层下一节点
			curHead = &next.NodeHeader
			prevHead[i] = curHead
		}
		// 此时下一节点前就是要插入的位置
		// 从最顶层开始跳过指向相同节点的层
		topLevel := curHead.levels[i]
		for i--; i >= 0 && curHead.levels[i] == topLevel; i-- {
			prevHead[i] = curHead
		}
	}

	// 建立前向指针
	if prev := prevHead[0]; prev != &list.NodeHeader {
		node.prev = prev.Element()
	}

	// 建立顶层
	if prev := prevHead[level-1]; prev != &list.NodeHeader {
		node.prevTopLevel = prev.Element()
	}

	// 每层都需写入指针, 指向与前一个最高节点指向相同
	for i := 0; i < level; i++ {
		node.levels[i] = prevHead[i].levels[i]
		prevHead[i].levels[i] = node
	}

	// 找下一节点最高层
	largestLevel := 0

	for i := level - 1; i >= 0; i-- {
		if node.levels[i] != nil {
			largestLevel = i + 1
			break
		}
	}

	// 调整下一节点指针前向指向
	if next := node.levels[0]; next != nil {
		next.prev = node
	}
	// 循环判断, 往后的层级比要插入的节点低的都要指向当前
	for i := 0; i < largestLevel; {
		next := node.levels[i]
		nextLevel := next.Level()

		if nextLevel <= level {
			next.prevTopLevel = node
		}

		i = nextLevel
	}

	// 记录尾节点
	if node.Next() == nil {
		list.back = node
	}

	list.length++
	return
}

// Find node with key
// if the key is not exist, return nil,err else node,nil
func (list *Skiplist) Find(key interface{}) (node *Node, err error) {
	if list.Empty() {
		return nil, err
	}

	score := list.calcScore(key)
	if score < list.Front().score || score > list.Back().score {
		return nil, err
	}

	prevHeader := &list.NodeHeader
	i := len(list.levels) - 1

	// 与插入时查找操作一致
	for i >= 0 {
		for next := prevHeader.levels[i]; next != nil; {
			if comp := list.compare(score, key, next); comp <= 0 {
				if comp == 0 {
					node = next
					return
				}
				break
			}
			prevHeader = &next.NodeHeader
			next = prevHeader.levels[i]
		}

		topLevel := prevHeader.levels[i]

		// Skip levels if they point to the same element as topLevel.
		for i--; i >= 0 && prevHeader.levels[i] == topLevel; i-- {
		}
	}
	return node, nil
}

// Erase 移除指定元素
// 返回删除元素的下一节点
func (list *Skiplist) Erase(key interface{}) (res *Node, err error) {
	node, err := list.Find(key)
	if node == nil || node.list != list {
		// 要删除的值不存在
		return nil, err
	}

	level := node.Level()

	// 需要找出所有前向节点
	max := 0 //
	prevNodes := make([]*Node, level)
	prev := node.prev

	for prev != nil && max < level {
		prevLevel := len(prev.levels)

		for ; max < prevLevel && max < level; max++ {
			prevNodes[max] = prev
		}

		for prev = prev.prevTopLevel; prev != nil && prev.Level() == prevLevel; prev = prev.prevTopLevel {
		}
	}

	// 修改前向节点的各层指针, 类似链表中cur->prev->next = cur->next
	for i := 0; i < max; i++ {
		prevNodes[i].levels[i] = node.levels[i]
	}
	// 如果要删除的节点最高,还要改变跳表头的指向
	for i := max; i < level; i++ {
		list.levels[i] = node.levels[i]
	}

	// 修改要删除节点各层下一个节点前向指向(cur->next->prev = cur->prev)
	if next := node.Next(); next != nil {
		next.prev = node.prev
	}

	for i := 0; i < level; {
		next := node.levels[i]

		if next == nil || next.prevTopLevel != node {
			break
		}

		i = next.Level()
		next.prevTopLevel = prevNodes[i-1]
	}

	// 更新back
	if list.back == node {
		list.back = node.prev
	}

	list.length--
	res = node.Next()
	node.Reset() // 清理节点
	return res, nil
}

// Update 更新值,若不存在直接返回
func (list *Skiplist) Update(key, value interface{}) (node *Node, err error) {
	node, err = list.Find(key)
	if node == nil {
		return nil, err
	}
	node.value = value
	return node, nil
}
func (list *Skiplist) Clear() {
	// 清空跳表
}
func (node *Node) Reset() {
	// 清理节点值
	node.list = nil
	node.prev = nil
	node.prevTopLevel = nil
	node.levels = nil
}

func (list *Skiplist) getRandomLevel() (level int) {
	level = 1
	// 随机数对2取模,奇偶各占一半 相当1/2概率加一层
	for (rand.Int31()&0xFFFF)%SkiplistP == 0 {
		level += 1
	}
	if level < SkiplistMaxlevel {
		return level
	} else {
		return SkiplistMaxlevel
	}
}

// 比较两种
func (list *Skiplist) compare(score float64, key interface{}, rhs *Node) int {
	if score != rhs.score {
		if score > rhs.score {
			return 1
		} else if score < rhs.score {
			return -1
		}

		return 0
	}

	return list.comparable.Compare(key, rhs.key)
}

func (list *Skiplist) calcScore(key interface{}) (score float64) {
	score = list.comparable.CalcScore(key)
	return
}
