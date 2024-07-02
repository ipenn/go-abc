package abc

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func TxGetOrderOneLock(tx *gorm.DB) Orders {
	var order Orders
	tx.Debug().Raw("SELECT * FROM orders WHERE (cmd >= 0 AND is_commission = 0) ORDER BY order_id ASC LIMIT 1").Scan(&order)
	if order.OrderId != 0 {
		tx.Debug().Exec("SELECT * FROM orders WHERE order_id = ? FOR UPDATE", order.OrderId)
	}
	return order
}

func GetOrder(where any) (order Orders) {
	db.Debug().Where(where).First(&order)
	return
}

func GetOrders(where string) (order []Orders) {
	db.Debug().Where(where).Order("close_time desc").Find(&order)
	return order
}

func TxSaveOrder(tx *gorm.DB, order Orders, update map[string]interface{}) error {
	if err := tx.Debug().Model(&order).Updates(update).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetAvailableVolume(where string) float64 {
	data := struct {
		Sum float64 `json:"sum"`
	}{}
	db.Debug().Select("sum(volume - spent) as sum").Model(Orders{}).Where(where).Scan(&data)
	return data.Sum
}

func GetAvailableVolumeOrders(where string) (data []Orders) {
	db.Debug().Model(Orders{}).Where(where).Order("close_time asc").Find(&data)
	return data
}

func UpdateOrderSpent(pay float64, order []Orders) int {
	var useAll []int
	var lastId int
	for _, o := range order {
		using := o.Volume - o.Spent
		if pay >= using {
			pay -= using
			useAll = append(useAll, o.OrderId)
		} else {
			pay += o.Spent
			lastId = o.OrderId
		}
	}
	if len(useAll) != 0 {
		if _, err := db.Debug().Raw("update orders set spent = volume where order_id in (?)", useAll).Rows(); err != nil {
			log.Println("abc UpdateOrderSpent1 ", err)
			return 0
		}
	}
	if lastId > 0 {
		if _, err := db.Debug().Raw("update orders set spent=? where order_id=?", pay, lastId).Rows(); err != nil {
			log.Println("abc UpdateOrderSpent2 ", err)
			return 0
		}
	}
	return 1
}

func GetOrderList(page, size int, where, order string) (count int64, orders []OrdersSimple, total Total) {
	db.Debug().Model(Orders{}).Where(where).Order(order).Limit(size).Offset((page - 1) * size).Scan(&orders)
	db.Debug().Model(Orders{}).Where(where).Count(&count)

	t, _ := SqlOperator(fmt.Sprintf(`SELECT
		IFNULL(sum(o.volume),0) as volume,
		IFNULL(sum(o.profit),0) as profit,
		IFNULL(sum(o.storage),0) as storage,
		IFNULL(sum(o.commission),0) as commission
FROM
	orders o
WHERE
	%s`, where))
	total.Volume = ToFloat64(PtoString(t, "volume"))
	total.Profit = ToFloat64(PtoString(t, "profit"))
	total.Storage = ToFloat64(PtoString(t, "storage"))
	total.Commission = ToFloat64(PtoString(t, "commission"))
	return count, orders, total
}

func GetOrderAllUserList(re bool, uid, page, size int, where, order string) (count int64, orders []OrdersSimple, total Total) {
	var redis RedisListData
	var result string

	if re {
		redisData := RDB.Get(Rctx, fmt.Sprintf("OrderAllUser-%d-%d", uid, size))
		result = redisData.Val()
	}

	if result == "" {
		db.Debug().Raw(fmt.Sprintf(`SELECT
	o.order_id,
	o.login,
	u.id as user_id,
	u.true_name,
	o.symbol,
	o.symbol_type,
	o.cmd,
	o.volume,
	o.open_price,
	o.open_time,
	o.close_price,
	o.close_time,
	o.sl,
	o.tp,
	o.profit,
	o.storage,
	o.commission,
	o.taxes,
	o.comment
FROM
	orders o
 	LEFT JOIN user u on SUBSTRING_INDEX(SUBSTRING_INDEX( user_id, ',',-2 ),',',1)=u.id
WHERE
	%s
	order by %s
limit %d,%d`, where, order, (page-1)*size, size)).Scan(&orders)
		operators, err := SqlOperator(fmt.Sprintf(`SELECT
		IFNULL(count(*),0) as count 
FROM
	orders o
WHERE
	%s`, where))
		if err != nil {
			log.Println("abc GetOrderAllUserList ", err)
		}
		count = ToInt64(PtoString(operators, "count"))

		t, err := SqlOperator(fmt.Sprintf(`SELECT
		IFNULL(sum(o.volume),0) as volume,
		IFNULL(sum(o.profit),0) as profit,
		IFNULL(sum(o.storage),0) as storage,
		IFNULL(sum(o.commission),0) as commission
FROM
	orders o
WHERE
	%s`, where))
		if err != nil {
			log.Println("abc GetOrderAllUserList ", err)
		}
		total.Volume = ToFloat64(PtoString(t, "volume"))
		total.Profit = ToFloat64(PtoString(t, "profit"))
		total.Storage = ToFloat64(PtoString(t, "storage"))
		total.Commission = ToFloat64(PtoString(t, "commission"))

		if re {
			redis.List = ToJSON(orders)
			redis.Count = count
			redis.Total = total
			str := ToJSON(redis)
			RDB.Set(Rctx, fmt.Sprintf("OrderAllUser-%d-%d", uid, size), str, 100*time.Second)
		}
	} else {
		json.Unmarshal([]byte(result), &redis)
		json.Unmarshal(redis.List, &orders)
		count = redis.Count
		total = redis.Total
	}
	return count, orders, total
}

func CustomerOrderInquiry(where, order string, page, size int) ([]interface{}, int64, interface{}) {
	res, _ := SqlOperators(fmt.Sprintf(`SELECT u.id, u.true_name, o.order_id, o.login, o.cmd, o.symbol, o.volume, o.open_time, o.open_price, o.close_time, o.close_price, o.sl, o.tp, o.profit, o.storage, o.commission, o.comment FROM orders o
					  				 LEFT JOIN user u
					  				 ON u.id = SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',',-2),',',1)
					  				 WHERE %v %v LIMIT %v,%v`, where, order, (page-1)*size, size))

	res1, _ := SqlOperator(fmt.Sprintf(`SELECT count(o.order_id) count FROM orders o
					  				 LEFT JOIN user u
					  				 ON u.id = SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',',-2),',',1)
					  				 WHERE %v`, where))

	res2, _ := SqlOperator(fmt.Sprintf(`SELECT sum(o.volume) volume, sum(o.profit) profit, sum(o.storage) storage, sum(o.commission) commission FROM orders o
					  				 LEFT JOIN user u
					  				 ON u.id = SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',',-2),',',1)
					  				 WHERE %v`, where))
	return res, ToInt64(PtoString(res1,"count")), res2
}
