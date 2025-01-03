---
categories:
- web3
date: 2024-09-24 21:06:02
showTableOfContents: 'true'
tags:
- Blockchain
title: 从 0 开始自己实现一个 Blockchain
type: post
---

# 从 0 开始自己实现一个 Blockchain

区块链的核心功能模块包括:
- 区块链基本数据结构, 区块的定义, header 的定义
- 共识算法, POW/POS....
- 数据库持久化存储区块数据
- 交易模型, 比特币 UTXO 模型或者以太坊的 Account 模型...
- 交易的生命周期


## 共识算法

## 持久化
### 数据库结构
> 详情可参考链接: https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_(ch_2):_Data_Storage

Bitcoin Core 使用两个 “bucket” 来存储数据：
-  blocks，它存储了描述一条链中所有块的元数据
- 另一个 bucket 是 chainstate，存储了一条链的状态，也就是当前所有的未花费的交易输出，和一些元数据
---

在 blocks 中，key -> value 为：

| Key | Value |
| --- | --- |
| `b` + 32 字节的 block hash | block index 结构 |
| `f` + 4 字节的 file number | file information |
| `l` + 4 字节的 file number | 最后一个 block file number |
| `R` + 1 字节的 boolean | 是否正在 reindex |
| `F` + 1 字节的 flag name length + flag name string | 1 byte boolean: various flags |
| `t` + 32 字节的 transaction hash | transaction index 结构 |

---
在 chainstate，key -> value 为：

| Key | Value |
| --- | --- |
| `c` + 32 字节的 transaction hash | unspent transaction output 结构 |
| `B` | 32 字节的 block hash: 存储区块链中第一个未花费交易输出的块 |


## 交易
### 交易模型

### 交易生命周期
以一个最简单的基于 UTXO 交易模型实现来说，大致为以下步骤:
1. 创世块里面包含了一个 coinbase 交易。在 coinbase 交易中，没有输入，所以也就不需要签名。coinbase 交易的输出包含了一个哈希过的公钥（使用的是 RIPEMD16(SHA256(PubKey)) 算法）

2. 当一个人发送币时，就会创建一笔交易。这笔交易的输入会引用之前交易的输出。每个输入会存储一个公钥（没有被哈希）和整个交易的一个签名。

3. 比特币网络中接收到交易的其他节点会对该交易进行验证。除了一些其他事情，他们还会检查：在一个输入中，公钥哈希与所引用的输出哈希相匹配（这保证了发送方只能花费属于自己的币）；签名是正确的（这保证了交易是由币的实际拥有者所创建）。

4. 当一个矿工准备挖一个新块时，他会将交易放到块中，然后开始挖矿。

5. 当新块被挖出来以后，网络中的所有其他节点会接收到一条消息，告诉其他人这个块已经被挖出并被加入到区块链。

6. 当一个块被加入到区块链以后，交易就算完成，它的输出就可以在新的交易中被引用

### 交易中的签名
早期比特币的签名过程:  ![bitcoinTx](https://3bcaf57.webp.li/myblog/bitcoinTx.png)

> 为什么需要复制当前交易(TxNew)到 TxCopy，然后清空 TxCopy 中所有输入脚本，将处理后的子脚本复制到待验证的输入脚本中?

- 隔离签名数据： 签名验证时需要验证的是交易的原始数据，而不包括签名本身. 如果直接使用包含签名的交易数据来验证签名，会导致循环依赖(因为签名是基于交易数据生成的),
      所以需要复制一份交易并清空输入脚本(因为输入脚本包含签名数据)


- 防止交易延展性： 如果不清空输入脚本，攻击者可以修改脚本内容(在不影响执行结果的情况下)
这样会产生不同的交易 ID，但实际执行结果相同, 这就是所谓的交易延展性问题. 通过清空输入脚本，确保相同的交易只能有一个确定的签名

#### 交易延展性
> 什么是交易延展性

同一笔交易可以有多个有效的表示形式, 这些不同形式会产生不同的交易ID(txid), 但它们在区块链上的执行结果是完全相同的

A. 签名延展性：
- ECDSA 签名中 s 值可以有两种形式：s 或 N-s (N是曲线阶数)
- 两种形式都是有效的签名
- 会产生不同的交易ID但验证结果相同

B. 脚本延展性：
- 在脚本中添加无效操作(如 OP_DROP OP_DROP)
- 使用不同的编码方式(如数字1可以编码为 0x01 或 0x0001)
- 添加额外的空操作(OP_NOP)

示例：
原始脚本：<signature> <pubkey> OP_CHECKSIG
可变形式：<signature> <pubkey> OP_NOP OP_CHECKSIG

> 交易延展性带来的主要问题

1. 双重花费风险
交易 ID 是确认交易唯一性的关键标识，如果它可以被修改，恶意用户可以将相同的交易以不同的交易 ID 提交多次，制造双重支付的混淆风险。尽管节点可能会最终识别到这是同一笔交易，但在实际应用中，接收方可能会面临无法及时确定交易有效性的问题。

2. 支付通道和多签交易的不稳定性
许多支付通道（如闪电网络）依赖于交易的唯一 ID 来追踪链上资金流动和链下交易状态。一旦交易 ID 被更改，支付通道中的资金可能会出现结算错误的风险，影响到闪电网络的效率和可靠性。

3. 复杂的交易追踪和错误处理
在多层协议或复杂应用中，交易 ID 的不确定性会导致交易追踪困难。应用程序往往依赖交易 ID 来确认资金流动，延展性问题会让接收方误以为某笔交易未完成或失效，从而需要额外的检查和确认。


#### 隔离见证
隔离见证（Segregated Witness，简称 SegWit）是比特币协议的一项升级，最早由比特币核心开发者 Pieter Wuille 提出，并在 2017 年正式实施，主要解决的就是交易的可扩展性和交易延展性问题
// TODO 隔离见证的代码实现

