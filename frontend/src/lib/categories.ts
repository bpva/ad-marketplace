export interface CategoryOption {
  slug: string;
  displayName: string;
}

export const ALL_CATEGORIES: CategoryOption[] = [
  { slug: "blogs", displayName: "Blogs" },
  { slug: "news_and_media", displayName: "News and media" },
  { slug: "humor_and_entertainment", displayName: "Humor and entertainment" },
  { slug: "technologies", displayName: "Technologies" },
  { slug: "economics", displayName: "Economics" },
  { slug: "business_and_startups", displayName: "Business and startups" },
  { slug: "cryptocurrencies", displayName: "Cryptocurrencies" },
  { slug: "travel", displayName: "Travel" },
  { slug: "marketing_pr_advertising", displayName: "Marketing, PR, advertising" },
  { slug: "psychology", displayName: "Psychology" },
  { slug: "design", displayName: "Design" },
  { slug: "politics", displayName: "Politics" },
  { slug: "art", displayName: "Art" },
  { slug: "law", displayName: "Law" },
  { slug: "education", displayName: "Education" },
  { slug: "books", displayName: "Books" },
  { slug: "linguistics", displayName: "Linguistics" },
  { slug: "career", displayName: "Career" },
  { slug: "edutainment", displayName: "Edutainment" },
  { slug: "courses_and_guides", displayName: "Courses and guides" },
  { slug: "sport", displayName: "Sport" },
  { slug: "fashion_and_beauty", displayName: "Fashion and beauty" },
  { slug: "medicine", displayName: "Medicine" },
  { slug: "health_and_fitness", displayName: "Health and fitness" },
  { slug: "pictures_and_photos", displayName: "Pictures and photos" },
  { slug: "software_and_applications", displayName: "Software and applications" },
  { slug: "video_and_films", displayName: "Video and films" },
  { slug: "music", displayName: "Music" },
  { slug: "games", displayName: "Games" },
  { slug: "food_and_cooking", displayName: "Food and cooking" },
  { slug: "quotes", displayName: "Quotes" },
  { slug: "handiwork", displayName: "Handiwork" },
  { slug: "family_and_children", displayName: "Family and children" },
  { slug: "nature", displayName: "Nature" },
  { slug: "interior_and_construction", displayName: "Interior and construction" },
  { slug: "telegram", displayName: "Telegram" },
  { slug: "instagram", displayName: "Instagram" },
  { slug: "sales", displayName: "Sales" },
  { slug: "transport", displayName: "Transport" },
  { slug: "religion", displayName: "Religion" },
  { slug: "esoterics", displayName: "Esoterics" },
  { slug: "darknet", displayName: "Darknet" },
  { slug: "bookmaking", displayName: "Bookmaking" },
  { slug: "shock_content", displayName: "Shock content" },
  { slug: "erotic", displayName: "Erotic" },
  { slug: "adult", displayName: "Adult" },
  { slug: "other", displayName: "Other" },
];

const CATEGORY_MAP = new Map(ALL_CATEGORIES.map((c) => [c.slug, c.displayName]));

export function getCategoryDisplayName(slug: string): string {
  return CATEGORY_MAP.get(slug) ?? slug;
}
