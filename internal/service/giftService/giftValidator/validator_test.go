package giftValidator

import (
	"gift-buyer/internal/config"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewGiftValidator(t *testing.T) {
	criterias := []config.Criterias{
		{MinPrice: 100, MaxPrice: 1000, TotalSupply: 50, Count: 5},
	}
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      false,
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator(criterias, giftParam)

	assert.NotNil(t, validator)
	assert.Equal(t, criterias, validator.criteria)
	assert.Equal(t, giftParam.TotalStarCap, validator.totalStarCap)
	assert.Equal(t, giftParam.TestMode, validator.testMode)
	assert.Equal(t, giftParam.LimitedStatus, validator.limitedStatus)
	assert.Equal(t, giftParam.ReleaseBy, validator.releaseBy)
}

func TestGiftValidator_IsEligible_SoldOut(t *testing.T) {
	criterias := []config.Criterias{
		{MinPrice: 100, MaxPrice: 1000, TotalSupply: 50, Count: 5},
	}
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      false,
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator(criterias, giftParam)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   500,
		SoldOut: true,
		Limited: true,
	}

	result, eligible := validator.IsEligible(gift)
	assert.False(t, eligible)
	assert.Nil(t, result)
}

func TestGiftValidator_IsEligible_LimitedStatusMismatch(t *testing.T) {
	criterias := []config.Criterias{
		{MinPrice: 100, MaxPrice: 1000, TotalSupply: 50, Count: 5},
	}
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      false,
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator(criterias, giftParam)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   500,
		SoldOut: false,
		Limited: false, // Mismatch with LimitedStatus: true
	}

	result, eligible := validator.IsEligible(gift)
	assert.False(t, eligible)
	assert.Nil(t, result)
}

func TestGiftValidator_IsEligible_ValidGift(t *testing.T) {
	criterias := []config.Criterias{
		{MinPrice: 100, MaxPrice: 1000, TotalSupply: 50, Count: 5, ReceiverType: []int{1}},
	}
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      true, // Enable test mode to bypass validations
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator(criterias, giftParam)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   500,
		SoldOut: false,
		Limited: true,
	}

	result, eligible := validator.IsEligible(gift)
	assert.True(t, eligible)
	assert.NotNil(t, result)
	assert.Equal(t, int64(5), result.CountForBuy)
	assert.Equal(t, []int{1}, result.ReceiverType)
}

func TestGiftValidator_IsEligible_TestMode(t *testing.T) {
	criterias := []config.Criterias{
		{MinPrice: 100, MaxPrice: 1000, TotalSupply: 50, Count: 5, ReceiverType: []int{1}},
	}
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      true, // Test mode enabled
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator(criterias, giftParam)

	gift := &tg.StarGift{
		ID:      1,
		Stars:   500,
		SoldOut: false,
		Limited: true,
	}

	result, eligible := validator.IsEligible(gift)
	assert.True(t, eligible)
	assert.NotNil(t, result)
}

func TestGiftValidator_PriceValid(t *testing.T) {
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      false,
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator([]config.Criterias{}, giftParam)

	criteria := config.Criterias{MinPrice: 100, MaxPrice: 1000}

	// Valid price
	gift := &tg.StarGift{Stars: 500}
	assert.True(t, validator.priceValid(criteria, gift))

	// Price too low
	gift = &tg.StarGift{Stars: 50}
	assert.False(t, validator.priceValid(criteria, gift))

	// Price too high
	gift = &tg.StarGift{Stars: 1500}
	assert.False(t, validator.priceValid(criteria, gift))

	// Edge cases
	gift = &tg.StarGift{Stars: 100} // Min price
	assert.True(t, validator.priceValid(criteria, gift))

	gift = &tg.StarGift{Stars: 1000} // Max price
	assert.True(t, validator.priceValid(criteria, gift))
}

func TestGiftValidator_SupplyValid_TestMode(t *testing.T) {
	giftParam := config.GiftParam{
		TotalStarCap:  10000,
		TestMode:      true, // Test mode bypasses supply validation
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator([]config.Criterias{}, giftParam)

	criteria := config.Criterias{TotalSupply: 50}
	gift := &tg.StarGift{Limited: true}

	// In test mode, supply validation should always pass
	assert.True(t, validator.supplyValid(criteria, gift))
}

func TestGiftValidator_StarCapValidation_TestMode(t *testing.T) {
	giftParam := config.GiftParam{
		TotalStarCap:  1000,
		TestMode:      true, // Test mode bypasses star cap validation
		LimitedStatus: true,
		ReleaseBy:     false,
	}

	validator := NewGiftValidator([]config.Criterias{}, giftParam)

	// Gift that would exceed star cap
	gift := &tg.StarGift{
		Stars:   500,
		Limited: true,
		Flags:   0,
	}

	// In test mode, star cap validation should always pass
	assert.True(t, validator.starCapValidation(gift))
}
