# 评论评价项目

##### 项目整体架构-CQRS

![image-20240723121850328](https://github.com/PanZhuo0/review/blob/main/pic/image-%E6%95%B4%E4%BD%93%E6%9E%B6%E6%9E%84.png)

外部用户可以通过API-gateway借助`review-c`访问`review-service`的服务

服务内部、`business`端、`operator`端可以通过grpc访问`review-service`的服务

- 查询流程以按商家ID查询为例

![image-20240723131910931](https://github.com/PanZhuo0/review/blob/main/pic/image-%E6%9F%A5%E8%AF%A2.png)

- CUD

![image-20240723132823478](https://github.com/PanZhuo0/review/blob/main/pic/image-CUD.png)

对评价的增删改操作将会直接操作MySQL数据库

- 数据一致性

![image-20240723133058888](https://github.com/PanZhuo0/review/blob/main/pic/image-%E4%B8%80%E8%87%B4%E6%80%A7.png)

1.通过canal从MySQL中获取binlog日志、发往Kafka

2.review-job从kafka对应主题中接受binlog日志的内容，进行解析，判断操作类型、将数据进行对应处理后，发往ES

##### 数据库表设计

```sql
CREATE TABLE review_info (
        `id` bigint(32) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
        `create_by` varchar(48) NOT NULL DEFAULT '' COMMENT '创建方标识',
        `update_by` varchar(48) NOT NULL DEFAULT '' COMMENT '更新方标识',
        `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        `delete_at` timestamp COMMENT '逻辑删除标记',
        `version`   int(10) unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁标记',

        `review_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '评价id',
        `content` varchar(512) NOT NULL COMMENT '评价内容',
        `score` tinyint(4) NOT NULL DEFAULT '0' COMMENT '评分',
        `service_score` tinyint(4) NOT NULL DEFAULT '0' COMMENT '商家服务评分',
        `express_score` tinyint(4) NOT NULL DEFAULT '0' COMMENT '物流评分',
        `has_media` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否有图或视频',
        `order_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '订单id',
        `sku_id` bigint(32) NOT NULL DEFAULT '0' COMMENT 'sku id',
        `spu_id` bigint(32) NOT NULL DEFAULT '0' COMMENT 'spu id',
        `store_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '店铺id',
        `user_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '用户id',
        `anonymous` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否匿名',
        `tags` varchar(1024) NOT NULL DEFAULT '' COMMENT '标签json',
        `pic_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：图片',
        `video_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：视频',
        `status` tinyint(4) NOT NULL DEFAULT '10' COMMENT '状态:10待审核；20审核通过；30审核不通过；40隐藏',
        `is_default` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否默认评价',
        `has_reply` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否有商家回复:0无;1有',
        `op_reason` varchar(512) NOT NULL DEFAULT '' COMMENT '运营审核拒绝原因',
        `op_remarks` varchar(512) NOT NULL DEFAULT '' COMMENT '运营备注',
        `op_user` varchar(64) NOT NULL DEFAULT '' COMMENT '运营者标识',

        `goods_snapshoot` varchar(2048) NOT NULL DEFAULT '' COMMENT '商品快照信息',
        `ext_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '信息扩展',
        `ctrl_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '控制扩展',
        PRIMARY KEY (`id`),
        KEY `idx_delete_at` (`delete_at`) COMMENT '逻辑删除索引',
        UNIQUE KEY `uk_review_id` (`review_id`) COMMENT '评价id索引',
        KEY `idx_order_id` (`order_id`) COMMENT '订单id索引',
        KEY `idx_user_id` (`user_id`) COMMENT '用户id索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评价表';

CREATE TABLE review_reply_info (
        `id` bigint(32) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
        `create_by` varchar(48) NOT NULL DEFAULT '' COMMENT '创建方标识',
        `update_by` varchar(48) NOT NULL DEFAULT '' COMMENT '更新方标识',
        `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        `delete_at` timestamp COMMENT '逻辑删除标记',
        `version` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁标记',

        `reply_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '回复id',
        `review_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '评价id',
        `store_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '店铺id',
        `content` varchar(512) NOT NULL COMMENT '评价内容',
        `pic_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：图片',
        `video_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：视频',

        `ext_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '信息扩展',
        `ctrl_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '控制扩展',
        PRIMARY KEY (`id`),
        KEY `idx_delete_at` (`delete_at`) COMMENT '逻辑删除索引',
        UNIQUE KEY `uk_reply_id` (`reply_id`) COMMENT '回复id索引',
        KEY `idx_review_id` (`review_id`) COMMENT '评价id索引',
        KEY `idx_store_id` (`store_id`) COMMENT '店铺id索引'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评价商家回复表';


CREATE TABLE review_appeal_info (
        `id` bigint(32) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
        `create_by` varchar(48) NOT NULL DEFAULT '' COMMENT '创建方标识',
        `update_by` varchar(48) NOT NULL DEFAULT '' COMMENT '更新方标识',
        `create_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
        `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
        `delete_at` timestamp COMMENT '逻辑删除标记',
        `version` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁标记',

        `appeal_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '回复id',
        `review_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '评价id',
        `store_id` bigint(32) NOT NULL DEFAULT '0' COMMENT '店铺id',
        `status` tinyint(4) NOT NULL DEFAULT '10' COMMENT '状态:10待审核；20申诉通过；30申诉驳回',
        `reason` varchar(255) NOT NULL COMMENT '申诉原因类别',
        `content` varchar(255) NOT NULL COMMENT '申诉内容描述',
        `pic_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：图片',
        `video_info` varchar(1024) NOT NULL DEFAULT '' COMMENT '媒体信息：视频',

        `op_remarks` varchar(512) NOT NULL DEFAULT '' COMMENT '运营备注',
        `op_user` varchar(64) NOT NULL DEFAULT '' COMMENT '运营者标识',

        `ext_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '信息扩展',
        `ctrl_json` varchar(1024) NOT NULL DEFAULT '' COMMENT '控制扩展',
        PRIMARY KEY (`id`),
        KEY `idx_delete_at` (`delete_at`) COMMENT '逻辑删除索引',
        KEY `idx_appeal_id` (`appeal_id`) COMMENT '申诉id索引',
        UNIQUE KEY `uk_review_id` (`review_id`) COMMENT '评价id索引',
        KEY `idx_store_id` (`store_id`) COMMENT '店铺id索引'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评价商家申诉表';
```

##### review-service提供的服务

```
1.用户创建评价
2.商家端可以对用户创建的评价进行申诉
3.运营端对用户评价、商家端对用户的评价进行审核
```

- .proto

```protobuf
syntax = "proto3";

package api.review.v1;

import "google/api/annotations.proto";
import "validate/validate.proto";

option go_package = "review-service/api/review/v1;v1";
option java_multiple_files = true;
option java_package = "api.review.v1";

// 定义评价服务
service Review {
	// C端创建评价
	rpc CreateReview (CreateReviewRequest) returns (CreateReviewReply){
		option (google.api.http) = {
			post: "/v1/review",
			body: "*"
		};
	};
	// C端获取评价详情
	rpc GetReview (GetReviewRequest) returns (GetReviewReply) {
		option (google.api.http) = {
			get: "/v1/review/{reviewID}"
		};
	}
	// O端审核评价
	rpc AuditReview (AuditReviewRequest) returns (AuditReviewReply) {
		option (google.api.http) = {
			post: "/v1/review/audit",
			body: "*"
		};
	}
	// B端回复评价
	rpc ReplyReview (ReplyReviewRequest) returns (ReplyReviewReply) {
		option (google.api.http) = {
			post: "/v1/review/reply",
			body: "*"
		};
	}
	// B端申诉评价
	rpc AppealReview (AppealReviewRequest) returns (AppealReviewReply) {
		option (google.api.http) = {
			post: "/v1/review/appeal",
			body: "*"
		};
	}
	// O端评价申诉审核
	rpc AuditAppeal (AuditAppealRequest) returns (AuditAppealReply) {
		option (google.api.http) = {
			post: "/v1/appeal/audit",
			body: "*"
		};
	}
	// C端查看userID下所有评价(使用ES)
	rpc ListReviewByUserID (ListReviewByUserIDRequest) returns (ListReviewByUserIDReply) {
		option (google.api.http) = {
			get: "/v1/{userID}/reviews",
		};
	}

	// 根据商家ID查询评价列表（分页）(使用ES)
	rpc ListReviewByStoreID (ListReviewByStoreIDRequest) returns (ListReviewByStoreIDReply) {}
}

message ListReviewByStoreIDRequest {
	int64 storeID = 1 [(validate.rules).int64 = {gt: 0}];
	int32 page = 2 [(validate.rules).int32= {gt: 0}];
	int32 size = 3 [(validate.rules).int32= {gt: 0}];
}

message ListReviewByStoreIDReply {
	repeated ReviewInfo list = 1;
}


// 创建评价的参数
message CreateReviewRequest {
	int64 userID = 1 [(validate.rules).int64 = {gt: 0}];
	int64 orderID = 2 [(validate.rules).int64 = {gt: 0}];
	int64 storeID = 3 [(validate.rules).int64 = {gt: 0}];
	int32 score = 4 [(validate.rules).int32 = {in: [1,2,3,4,5]}];
	int32 serviceScore = 5 [(validate.rules).int32 = {in: [1,2,3,4,5]}];
	int32 expressScore = 6 [(validate.rules).int32 = {in: [1,2,3,4,5]}];
	string content = 7 [(validate.rules).string = {min_len: 8, max_len: 255}];
	string picInfo = 8;
	string videoInfo = 9;
	bool anonymous = 10;
}
// 创建评价的回复
message CreateReviewReply {
	int64 reviewID = 1;
}

// 获取评价详情的请求参数
message GetReviewRequest {
	int64 reviewID = 1 [(validate.rules).int64 = {gt: 0}];
}

// 获取评价详情的响应
message GetReviewReply {
	ReviewInfo data = 1;
}

// 评价信息
message ReviewInfo{
	int64 reviewID = 1;
	int64 userID = 2;
	int64 orderID = 3;
	int32 score = 4;
	int32 serviceScore = 5;
	int32 expressScore = 6;
	string content = 7;
	string picInfo = 8;
	string videoInfo = 9;
	int32 status = 10;
}

// 审核评价的请求
message AuditReviewRequest {
	int64 reviewID = 1 [(validate.rules).int64 = {gt: 0}];
	int32 status = 2 [(validate.rules).int32 = {gt: 0}];
	string opUser = 3 [(validate.rules).string = {min_len: 2}];
	string opReason = 4 [(validate.rules).string = {min_len: 2}];
	optional string opRemarks = 5;
}

// 审核评价的返回值
message AuditReviewReply {
	int64 reviewID = 1;
	int32 status = 2;
}

// 回复评价的请求
message ReplyReviewRequest{
	int64 reviewID = 1 [(validate.rules).int64 = {gt: 0}];
	int64 storeID = 2 [(validate.rules).int64 = {gt: 0}];
	string content = 3 [(validate.rules).string = {min_len: 2, max_len:200}];
	string picInfo = 4;
	string videoInfo = 5;
}

// 回复评价的返回值
message ReplyReviewReply{
	int64 replyID = 1;
}

// AppealReviewRequest 申诉评价的请求参数
message AppealReviewRequest{
	int64 reviewID = 1 [(validate.rules).int64 = {gt: 0}];
	int64 storeID = 2 [(validate.rules).int64 = {gt: 0}];
	string reason = 3 [(validate.rules).string = {min_len: 2, max_len:200}];
	string content = 4 [(validate.rules).string = {min_len: 2, max_len:200}];
	string picInfo = 5;
	string videoInfo = 6;
}

// 对评价进行申诉的返回值
message AppealReviewReply{
    int64 appealID = 1;
}

// 对申诉进行审核的请求
message AuditAppealRequest{
	int64 appealID = 1 [(validate.rules).int64 = {gt: 0}];
	int64 reviewID = 2 [(validate.rules).int64 = {gt: 0}];
	int32 status = 3 [(validate.rules).int32 = {gt: 0}];
	string opUser = 4 [(validate.rules).string = {min_len: 2}];
	optional string opRemarks = 5;
}

// 对申诉进行审核的返回值
message AuditAppealReply{
}

// 用户评价列表的请求
message ListReviewByUserIDRequest{
	int64 userID = 1 [(validate.rules).int64 = {gt: 0}];
	int32 page = 2 [(validate.rules).int32 = {gt: 0}];
	int32 size = 3 [(validate.rules).int32 = {gt: 0}];
}

// 用户评价列表的返回值
message ListReviewByUserIDReply{
	repeated ReviewInfo list = 1;
}
```

对应的实现在`review-service/internal`

##### 环境

- 使用到的组件

```
ES
	用于数据的各种条件下的检索
MySQL
	对数据进行存储
Redis
	对查询结果做缓存key，会定时删除
Canal
	将MySQL的binlog日志数据发往kafka，然后交给ES处理
Consul
	服务注册与服务发现
```

