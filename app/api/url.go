package api

import "strconv"

func OfferUrl(baseUrl string, offer Offer) string {
	return baseUrl + "/oferty/" + offer.Vendor.Slug + "/" + offer.Slug + "-" + strconv.FormatInt(offer.Id, 10)
}
