package context

func IsApproved(item ApprovedItem) bool { return item.State == StateApproved }
