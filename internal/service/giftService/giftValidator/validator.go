// Package giftValidator provides gift validation functionality for the gift buying system.
// It implements criteria-based validation to determine which gifts are eligible for purchase
// based on price ranges, supply constraints, and total spending limits.
package giftValidator

import (
	"gift-buyer/internal/config"
	"gift-buyer/internal/service/giftService/giftTypes"

	"github.com/gotd/td/tg"
)

// GiftValidator implements the GiftValidator interface for validating gifts
// against configured purchase criteria. It evaluates gifts based on price,
// supply availability, and total star spending caps.
type giftValidatorImpl struct {
	limitedStatus, releaseBy, premium bool
	// criteria contains the list of validation criteria for gift purchases
	criteria []config.Criterias

	// totalStarCap is the maximum total stars that can be spent across all gifts
	totalStarCap int64

	// testMode enables test mode which bypasses certain validations
	testMode bool
}

// NewGiftValidator creates a new GiftValidator instance with the specified criteria.
// The validator will use the provided criteria to evaluate gift eligibility.
//
// Parameters:
//   - criterias: slice of criteria defining price ranges, supply limits, and purchase counts
//   - totalStarCap: maximum total stars that can be spent across all eligible gifts
//   - testMode: if true, bypasses supply and star cap validations for testing
//
// Returns:
//   - giftInterfaces.GiftValidator: configured gift validator instance
func NewGiftValidator(criterias []config.Criterias, giftParam config.GiftParam) *giftValidatorImpl {
	return &giftValidatorImpl{
		criteria:      criterias,
		totalStarCap:  giftParam.TotalStarCap,
		premium:       giftParam.OnlyPremium,
		testMode:      giftParam.TestMode,
		limitedStatus: giftParam.LimitedStatus,
		releaseBy:     giftParam.ReleaseBy,
	}
}

// IsEligible checks if a gift meets any of the configured purchase criteria.
// It evaluates the gift against all criteria and returns the purchase count
// for the first matching criteria.
//
// The validation process checks:
//   - Gift is not sold out
//   - Price falls within configured range
//   - Supply meets minimum requirements (unless in test mode)
//   - Total star cap is not exceeded (unless in test mode)
//
// Parameters:
//   - gift: the star gift to validate against criteria
//
// Returns:
//   - int64: number of gifts to purchase if eligible (0 if not eligible)
//   - bool: true if the gift meets any criteria, false otherwise
func (gv *giftValidatorImpl) IsEligible(gift *tg.StarGift) (*giftTypes.GiftRequire, bool) {
	if gift.SoldOut {
		return nil, false
	}

	if gift.Limited != gv.limitedStatus {
		return nil, false
	}

	if !gv.releaseByValidation(gift) {
		return nil, false
	}

	if ok := gv.premiumValidation(gift); !ok {
		return nil, false
	}

	for _, criteria := range gv.criteria {
		if gv.priceValid(criteria, gift) && gv.supplyValid(criteria, gift) && gv.starCapValidation(gift) {
			return &giftTypes.GiftRequire{
				Gift:         gift,
				ReceiverType: criteria.ReceiverType,
				CountForBuy:  criteria.Count,
				Hide:         criteria.Hide,
			}, true
		}
	}

	return nil, false
}

// priceValid checks if the gift price falls within the specified criteria range.
//
// Parameters:
//   - criteria: the criteria containing min and max price limits
//   - gift: the star gift to validate
//
// Returns:
//   - bool: true if the gift price is within the criteria range
func (gv *giftValidatorImpl) priceValid(criteria config.Criterias, gift *tg.StarGift) bool {
	giftPrice := gift.GetStars()
	if giftPrice >= criteria.MinPrice && giftPrice <= criteria.MaxPrice {
		return true
	}

	return false
}

// supplyValid checks if the gift supply meets the minimum requirements.
// In test mode, this validation is bypassed and always returns true.
//
// For limited gifts, it checks:
//   - Gift is not sold out
//   - Remaining supply is greater than 0
//   - Total supply is not greater than the maximum allowed supply
//
// For unlimited gifts, it always returns true.
//
// Parameters:
//   - criteria: the criteria containing total supply requirements
//   - gift: the star gift to validate
//
// Returns:
//   - bool: true if the gift supply meets requirements
func (gv *giftValidatorImpl) supplyValid(criteria config.Criterias, gift *tg.StarGift) bool {
	if gv.testMode {
		return true
	}

	if gift.Limited {
		remains, hasRemains := gift.GetAvailabilityRemains()
		if !hasRemains || remains <= 0 {
			return false
		}

		totalSupply, hasTotalSupply := gift.GetAvailabilityTotal()
		if !hasTotalSupply {
			return false
		}

		if int64(totalSupply) <= criteria.TotalSupply {
			return true
		}
		return false
	}

	return true
}

// starCapValidation checks if purchasing the gift would exceed the total star spending cap.
// In test mode, this validation is bypassed and always returns true.
//
// The validation calculates the total cost as: gift_price * total_supply
// and ensures it doesn't exceed the configured total star cap.
//
// Parameters:
//   - gift: the star gift to validate
//
// Returns:
//   - bool: true if the gift doesn't exceed the star spending cap
func (gv *giftValidatorImpl) starCapValidation(gift *tg.StarGift) bool {
	if gv.testMode {
		return true
	}

	price := gift.GetStars()
	giftSupply, _ := gift.GetAvailabilityTotal()
	return (price * int64(giftSupply)) <= gv.totalStarCap
}

func (gv *giftValidatorImpl) releaseByValidation(gift *tg.StarGift) bool {
	_, hasReleasedBy := gift.GetReleasedBy()

	if hasReleasedBy && !gv.releaseBy {
		return false
	}
	return true
}

func (gv *giftValidatorImpl) premiumValidation(gift *tg.StarGift) bool {
	premium := gift.GetRequirePremium()
	if gv.premium && !premium {
		return false
	}

	return true
}
