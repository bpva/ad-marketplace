package entity

type Category struct {
	ID          int    `db:"id" json:"id"`
	Slug        string `db:"slug" json:"slug"`
	DisplayName string `db:"display_name" json:"display_name"`
}

type ChannelCategory string

const (
	CategoryBlogs                  ChannelCategory = "blogs"
	CategoryNewsAndMedia           ChannelCategory = "news_and_media"
	CategoryHumorAndEntertainment  ChannelCategory = "humor_and_entertainment"
	CategoryTechnologies           ChannelCategory = "technologies"
	CategoryEconomics              ChannelCategory = "economics"
	CategoryBusinessAndStartups    ChannelCategory = "business_and_startups"
	CategoryCryptocurrencies       ChannelCategory = "cryptocurrencies"
	CategoryTravel                 ChannelCategory = "travel"
	CategoryMarketingPRAdvertising ChannelCategory = "marketing_pr_advertising"
	CategoryPsychology             ChannelCategory = "psychology"
	CategoryDesign                 ChannelCategory = "design"
	CategoryPolitics               ChannelCategory = "politics"
	CategoryArt                    ChannelCategory = "art"
	CategoryLaw                    ChannelCategory = "law"
	CategoryEducation              ChannelCategory = "education"
	CategoryBooks                  ChannelCategory = "books"
	CategoryLinguistics            ChannelCategory = "linguistics"
	CategoryCareer                 ChannelCategory = "career"
	CategoryEdutainment            ChannelCategory = "edutainment"
	CategoryCoursesAndGuides       ChannelCategory = "courses_and_guides"
	CategorySport                  ChannelCategory = "sport"
	CategoryFashionAndBeauty       ChannelCategory = "fashion_and_beauty"
	CategoryMedicine               ChannelCategory = "medicine"
	CategoryHealthAndFitness       ChannelCategory = "health_and_fitness"
	CategoryPicturesAndPhotos      ChannelCategory = "pictures_and_photos"
	CategorySoftwareAndApps        ChannelCategory = "software_and_applications"
	CategoryVideoAndFilms          ChannelCategory = "video_and_films"
	CategoryMusic                  ChannelCategory = "music"
	CategoryGames                  ChannelCategory = "games"
	CategoryFoodAndCooking         ChannelCategory = "food_and_cooking"
	CategoryQuotes                 ChannelCategory = "quotes"
	CategoryHandiwork              ChannelCategory = "handiwork"
	CategoryFamilyAndChildren      ChannelCategory = "family_and_children"
	CategoryNature                 ChannelCategory = "nature"
	CategoryInteriorAndConstr      ChannelCategory = "interior_and_construction"
	CategoryTelegram               ChannelCategory = "telegram"
	CategoryInstagram              ChannelCategory = "instagram"
	CategorySales                  ChannelCategory = "sales"
	CategoryTransport              ChannelCategory = "transport"
	CategoryReligion               ChannelCategory = "religion"
	CategoryEsoterics              ChannelCategory = "esoterics"
	CategoryDarknet                ChannelCategory = "darknet"
	CategoryBookmaking             ChannelCategory = "bookmaking"
	CategoryShockContent           ChannelCategory = "shock_content"
	CategoryErotic                 ChannelCategory = "erotic"
	CategoryAdult                  ChannelCategory = "adult"
	CategoryOther                  ChannelCategory = "other"
)

var AllCategories = map[ChannelCategory]struct{}{
	CategoryBlogs:                  {},
	CategoryNewsAndMedia:           {},
	CategoryHumorAndEntertainment:  {},
	CategoryTechnologies:           {},
	CategoryEconomics:              {},
	CategoryBusinessAndStartups:    {},
	CategoryCryptocurrencies:       {},
	CategoryTravel:                 {},
	CategoryMarketingPRAdvertising: {},
	CategoryPsychology:             {},
	CategoryDesign:                 {},
	CategoryPolitics:               {},
	CategoryArt:                    {},
	CategoryLaw:                    {},
	CategoryEducation:              {},
	CategoryBooks:                  {},
	CategoryLinguistics:            {},
	CategoryCareer:                 {},
	CategoryEdutainment:            {},
	CategoryCoursesAndGuides:       {},
	CategorySport:                  {},
	CategoryFashionAndBeauty:       {},
	CategoryMedicine:               {},
	CategoryHealthAndFitness:       {},
	CategoryPicturesAndPhotos:      {},
	CategorySoftwareAndApps:        {},
	CategoryVideoAndFilms:          {},
	CategoryMusic:                  {},
	CategoryGames:                  {},
	CategoryFoodAndCooking:         {},
	CategoryQuotes:                 {},
	CategoryHandiwork:              {},
	CategoryFamilyAndChildren:      {},
	CategoryNature:                 {},
	CategoryInteriorAndConstr:      {},
	CategoryTelegram:               {},
	CategoryInstagram:              {},
	CategorySales:                  {},
	CategoryTransport:              {},
	CategoryReligion:               {},
	CategoryEsoterics:              {},
	CategoryDarknet:                {},
	CategoryBookmaking:             {},
	CategoryShockContent:           {},
	CategoryErotic:                 {},
	CategoryAdult:                  {},
	CategoryOther:                  {},
}
