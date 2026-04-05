package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
)

func Test_IfGetTokenWithBadGrantType_ShouldReturnBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.POST("/v1/token").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithBytes([]byte("grant_type=authorization_code&username=tst&password=tst")).
		Expect().
		Status(http.StatusBadRequest)
}

func Test_IfValidGetToken_ShouldReturnToken(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.POST("/v1/token").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithBytes([]byte("grant_type=password&scope=read+write&username=tst&password=tst")).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Keys().
		ContainsAll("access_token", "token_type")
}

func Test_IfRegisterWithValidBody_ShouldCreateUser(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	suffix := time.Now().UnixNano()

	e.POST("/v1/register").
		WithJSON(map[string]interface{}{
			"login":     fmt.Sprintf("user_register_%d", suffix),
			"email":     fmt.Sprintf("user_register_%d@test.com", suffix),
			"password":  "123456",
			"firstname": "Ivan",
			"lastname":  "Ivanov",
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
}

func Test_IfRegisterWithoutLogin_ShouldReturnBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	obj := e.POST("/v1/register").
		WithJSON(map[string]interface{}{
			"login":     "",
			"email":     "user_register_2@test.com",
			"password":  "123456",
			"firstname": "Ivan",
			"lastname":  "Ivanov",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object()

	obj.Value("code").String().IsEqual("BadRequest")
	obj.Value("message").String().IsEqual("login is required")
}

func Test_IfGetUserWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/user").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfGetUserWithValidToken_ShouldReturnUser(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	resp := e.GET("/v1/user").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	resp.ContainsKey("firstname")
	resp.ContainsKey("lastname")
	resp.ContainsKey("hasPremium")
}

func Test_IfGetProductsWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/products").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfGetProductsWithValidToken_ShouldReturnProducts(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	resp := e.GET("/v1/products").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	resp.ContainsKey("products")
	resp.ContainsKey("totalCount")
}

func Test_IfGetProductByIDWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/product/1").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfGetProductByIDWithInvalidID_ShouldReturnForbidden(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	e.GET("/v1/product/abc").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusForbidden)
}

func Test_IfGetCartWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/cart").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfReplaceCartWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.PUT("/v1/cart").
		WithJSON(map[string]interface{}{
			"products": []map[string]interface{}{
				{
					"id":       1,
					"quantity": 1,
				},
			},
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfReplaceCartWithInvalidBody_ShouldReturnBadRequestOrForbidden(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	e.PUT("/v1/cart").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"products": []map[string]interface{}{
				{
					"id":       0,
					"quantity": 1,
				},
			},
		}).
		Expect().
		Status(http.StatusBadRequest)
}

func Test_IfCreateOrderWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.POST("/v1/order").
		WithJSON(map[string]interface{}{
			"address": "Moscow, Lenina 1",
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfCreateOrderWithoutAddress_ShouldReturnBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	obj := e.POST("/v1/order").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"address": "",
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object()

	obj.Value("code").String().IsEqual("BadRequest")
	obj.Value("message").String().IsEqual("address is required")
}

func Test_IfGetOrdersWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/orders").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfGetOrderWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.GET("/v1/order/1").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfGetOrderWithInvalidID_ShouldReturnForbidden(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	e.GET("/v1/order/abc").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusForbidden)
}

func Test_IfDeleteOrderWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.DELETE("/v1/order/1").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfDeleteOrderWithInvalidID_ShouldReturnForbiddenOrBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	e.DELETE("/v1/order/abc").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusForbidden)
}

func Test_IfPayWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.POST("/v1/pay").
		WithJSON(map[string]interface{}{
			"paymentType": "premium",
			"amount":      100,
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfPayWithInvalidPaymentType_ShouldReturnBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	obj := e.POST("/v1/pay").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"paymentType": "unknown",
			"amount":      100,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object()

	obj.Value("code").String().IsEqual("BadRequest")
	obj.Value("message").String().IsEqual("invalid paymentType")
}

func Test_IfPayPremiumWithInvalidAmount_ShouldReturnBadRequest(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	obj := e.POST("/v1/pay").
		WithHeader("Authorization", "Bearer "+token).
		WithJSON(map[string]interface{}{
			"paymentType": "premium",
			"amount":      0,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object()

	obj.Value("code").String().IsEqual("BadRequest")
	obj.Value("message").String().IsEqual("invalid amount")
}

func Test_IfCreateProductsWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.POST("/v1/products").
		WithJSON([]map[string]interface{}{
			{
				"name":  "test product",
				"price": 100,
			},
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfUpdateProductsWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.PUT("/v1/products").
		WithJSON([]map[string]interface{}{
			{
				"id":    1,
				"name":  "updated",
				"price": 200,
			},
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfDeleteProductWithoutToken_ShouldReturnUnauthorized(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)

	e.DELETE("/v1/product/1").
		Expect().
		Status(http.StatusUnauthorized)
}

func Test_IfDeleteProductWithInvalidID_ShouldReturnForbidden(t *testing.T) {
	e := httpexpect.Default(t, testServerURL)
	token := getTestToken(t)

	e.DELETE("/v1/product/abc").
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusForbidden)
}