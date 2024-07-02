package abc

import "strings"

func (group SalesGroup) AllSalesGroups(users []User) []User {
	if len(users) == 0 {
		return group
	}
	me := users[0]
	users = users[1:]
	users = append(users, User{})
	meOwn, lUsers := SalesGroup([]User{}).SalesGroups(me, users)
	meOwn = append([]User{me}, meOwn...)
	group = append(group, meOwn...)
	group = group.AllSalesGroups(CutNilTail(lUsers))
	return group
}

func CutNilTail(users []User) []User {
	var index int
	for i, user := range users {
		if user.Id == 0 {
			index = i
			break
		}
	}

	return users[:index]
}

type SalesGroup []User

func (group SalesGroup) SalesGroups(me User, users []User) (SalesGroup, []User) {
	if len(users) == 0 {
		return group, users
	}

	var sale User
	var i []int
	for index, user := range users {
		if user.UserType == "sales" && user.ParentId == me.Id {
			sale = user
			group = append(group, user)
			i = append(i, index)
			break
		}
	}

	if sale.Id == 0 {
		return group, users
	}

	lenSale := len(strings.Split(sale.Path, ","))
	recurUser := make([]User, 0)
	for index, user := range users {
		lenU := len(strings.Split(user.Path, ","))
		if user.SalesId == sale.Id && (lenU-1 == lenSale) {
			group = append(group, user)
			recurUser = append(recurUser, user)
			i = append(i, index)
		}
		if user.SalesId == sale.Id && (lenU-2 == lenSale) && user.UserType != "sales" {
			group = append(group, user)
			recurUser = append(recurUser, user)
			i = append(i, index)
		}
	}

	var cnt int
	for _, j := range i {
		j -= cnt
		users = append(users[0:j], users[j+1:]...)
		cnt++
	}

	for _, user := range recurUser {
		group, _ = group.SalesGroups(user, users)
	}
	group, _ = group.SalesGroups(me, users)

	return group, users
}

func NodeGroups(users []User) []User {
	me := users[0]
	users = users[1:]
	users = append(users, User{})
	nodes := SalesGroup([]User{}).NodeGroups(me, users)
	nodes = append([]User{me}, nodes...)
	return nodes
}

func (group SalesGroup) NodeGroups(me User, users []User) SalesGroup {
	if len(users) == 0 {
		return group
	}

	var i []int
	recurUser := make([]User, 0)
	for j, user := range users {
		if user.ParentId == me.Id {
			group = append(group, user)
			i = append(i, j)
			recurUser = append(recurUser, user)
			break
		}
	}

	if len(recurUser) == 0 {
		return group
	}

	var cnt int
	for _, j := range i {
		j -= cnt
		users = append(users[0:j], users[j+1:]...)
		cnt++
	}

	for _, user := range recurUser {
		group = group.NodeGroups(user, users)
	}
	group = group.NodeGroups(me, users)

	return group
}
