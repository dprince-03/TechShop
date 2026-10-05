package sales

import (
	"strconv"

	"github.com/dprince-03/techshop/backend/internal/httpx"
)

func couponMinOrder(min int64) error {
	msg := "Spend at least ₦" + strconv.FormatInt(min/100, 10) + " to use this coupon."
	return httpx.Invalid("coupon_min_order", msg).Field("body.couponCode", msg, nil)
}

func couponNotApplicable() error {
	msg := "This coupon doesn't apply to anything in your cart."
	return httpx.Invalid("coupon_not_applicable", msg).Field("body.couponCode", msg, nil)
}
