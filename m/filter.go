package m

// 权限过滤器
type Filter struct {
	set map[string][]uint
}

const (
	// 权限增量值字面量, 根据需求可以随意追加，
	// 数据库存储只需要存储这几个字面量即可
	// 随项目启动会重新初始化
	USER = iota + 1
	SALES1
	IB1
	SALES2
	IB2
	SALES3
	IB3
	ADMIN
)

var (
	Keys = []string{
		// 基本权限
		"user",
		"sale1",
		"ib1",
		"sale2",
		"ib2",
		"sale3",
		"ib3",
		"admin",
		//组合权限
		"sale2sale3", // key值主要为了满足开发人员易读,自由取名
		"adminsale1sale2sale3",
		"sale1sale2sale3",
		"ib1ib2ib3",
		"userib1ib2ib3",
		"adminib1ib2ib3",
		"ib2ib3",
		"sale1sale2sale3ib1ib2ib3",
		"ib2ib3adminuser",
		"ib2ib3sale2sale3admin",
	}
	//随项目启动初始化
	filter *Filter
)

const (
	// 与Keys顺序对应，Keys取数组主要为了方便统计后续地len长度，
	// 方便动态计算容器空间

	// 基本权限
	User = iota
	Sale1
	Ib1
	Sale2
	Ib2
	Sale3
	Ib3
	Admin
	Sale2Sale3 // 组合权限
	AdminSale1Sale2Sale3Ib1Ib2Ib3
	Sale1Sale2Sale3
	Ib1Ib2Ib3
	UserIb1Ib2Ib3
	AdminIb1Ib2Ib3
	Ib2Ib3
	Sale1Sale2Sale3Ib1Ib2Ib3
	Ib1Ib2Ib3AdminUser
	Ib2Ib3Sale2Sale3Admin
)

func init() {
	filter = &Filter{
		set: make(map[string][]uint, len(Keys)),
	}

	// 手动配置权限，项目启动便初始化
	// ----------基本权限------------
	filter.addRole(Keys[User], USER)
	filter.addRole(Keys[Sale1], SALES1)
	filter.addRole(Keys[Sale2], SALES2)
	filter.addRole(Keys[Sale3], SALES3)
	filter.addRole(Keys[Ib1], IB1)
	filter.addRole(Keys[Ib2], IB2)
	filter.addRole(Keys[Ib3], IB3)
	filter.addRole(Keys[Admin], ADMIN)
	filter.addRole(Keys[Sale2Sale3], SALES2, SALES3)
	filter.addRole(Keys[AdminSale1Sale2Sale3Ib1Ib2Ib3], ADMIN, SALES1, SALES2, SALES3, IB1, IB2, IB3)
	filter.addRole(Keys[Sale1Sale2Sale3], SALES1, SALES2, SALES3)
	filter.addRole(Keys[Ib1Ib2Ib3], IB1, IB2, IB3)
	filter.addRole(Keys[UserIb1Ib2Ib3], USER, IB1, IB2, IB3)
	filter.addRole(Keys[AdminIb1Ib2Ib3], ADMIN, IB1, IB2, IB3)
	filter.addRole(Keys[Ib2Ib3], IB2, IB3)
	filter.addRole(Keys[Sale1Sale2Sale3Ib1Ib2Ib3], SALES1, SALES2, SALES3, IB1, IB2, IB3)
	filter.addRole(Keys[Ib1Ib2Ib3AdminUser], IB1, IB2, IB3, ADMIN, USER)
	filter.addRole(Keys[Ib2Ib3Sale2Sale3Admin], IB2, IB3, SALES2, SALES3, ADMIN)
	// ----------组合权限-------------
}

func (f *Filter) addRole(key string, roles ...int) {
	set := make([]uint, len(Keys)/8+1)
	for _, role := range roles {
		f.set[key] = set
		f.set[key][role>>3] |= 1 << (role & 0x07)
	}
}

func (f *Filter) verify(key string, role int) bool {
	if f.set[key][role>>3]&(1<<(role&0x07)) == 0 {
		return false
	}
	return true
}
