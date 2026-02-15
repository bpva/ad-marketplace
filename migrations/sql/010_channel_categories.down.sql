DROP MATERIALIZED VIEW channel_marketplace;

DROP TABLE channel_categories;
DROP TABLE categories;

CREATE MATERIALIZED VIEW channel_marketplace AS
SELECT
    c.id AS channel_id,
    c.telegram_channel_id,
    c.title,
    c.username,
    c.photo_small_file_id,
    c.photo_big_file_id,
    COALESCE(ci.about, '') AS about,
    ci.subscribers,
    ci.linked_chat_id,
    ci.languages,
    ci.top_hours,
    ci.reactions_by_emotion,
    ci.story_reactions_by_emotion,
    ci.recent_posts,
    (
        SELECT jsonb_agg(jsonb_build_object(
            'id', caf.id,
            'channel_id', caf.channel_id,
            'format_type', caf.format_type,
            'is_native', caf.is_native,
            'feed_hours', caf.feed_hours,
            'top_hours', caf.top_hours,
            'price_nano_ton', caf.price_nano_ton,
            'created_at', caf.created_at
        ) ORDER BY caf.created_at)
        FROM channel_ad_formats caf
        WHERE caf.channel_id = c.id
    ) AS ad_formats,
    (
        SELECT CASE WHEN COUNT(*) >= 1
            THEN (SUM(vbs.val::bigint) / COUNT(*))::int
            ELSE NULL END
        FROM channel_historical_stats chs,
            jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
        WHERE chs.channel_id = c.id
            AND chs.date = CURRENT_DATE - INTERVAL '1 day'
    ) AS avg_daily_views_1d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN (SUM(vbs.val::bigint) / COUNT(DISTINCT chs.date))::int
            ELSE NULL END
        FROM channel_historical_stats chs,
            jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '7 days'
    ) AS avg_daily_views_7d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN (SUM(vbs.val::bigint) / COUNT(DISTINCT chs.date))::int
            ELSE NULL END
        FROM channel_historical_stats chs,
            jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '30 days'
    ) AS avg_daily_views_30d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN SUM(vbs.val::bigint)::int
            ELSE NULL END
        FROM channel_historical_stats chs,
            jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '7 days'
    ) AS total_views_7d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN SUM(vbs.val::bigint)::int
            ELSE NULL END
        FROM channel_historical_stats chs,
            jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '30 days'
    ) AS total_views_30d,
    (
        SELECT CASE WHEN COUNT(*) >= 2
            THEN (
                (SELECT (chs2.data->>'subscribers')::int
                 FROM channel_historical_stats chs2
                 WHERE chs2.channel_id = c.id
                     AND chs2.date >= CURRENT_DATE - INTERVAL '7 days'
                 ORDER BY chs2.date DESC LIMIT 1)
                -
                (SELECT (chs3.data->>'subscribers')::int
                 FROM channel_historical_stats chs3
                 WHERE chs3.channel_id = c.id
                     AND chs3.date >= CURRENT_DATE - INTERVAL '7 days'
                 ORDER BY chs3.date ASC LIMIT 1)
            )
            ELSE NULL END
        FROM channel_historical_stats chs
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '7 days'
    ) AS sub_growth_7d,
    (
        SELECT CASE WHEN COUNT(*) >= 2
            THEN (
                (SELECT (chs2.data->>'subscribers')::int
                 FROM channel_historical_stats chs2
                 WHERE chs2.channel_id = c.id
                     AND chs2.date >= CURRENT_DATE - INTERVAL '30 days'
                 ORDER BY chs2.date DESC LIMIT 1)
                -
                (SELECT (chs3.data->>'subscribers')::int
                 FROM channel_historical_stats chs3
                 WHERE chs3.channel_id = c.id
                     AND chs3.date >= CURRENT_DATE - INTERVAL '30 days'
                 ORDER BY chs3.date ASC LIMIT 1)
            )
            ELSE NULL END
        FROM channel_historical_stats chs
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '30 days'
    ) AS sub_growth_30d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN (SUM((chs.data->>'interactions')::bigint) / COUNT(DISTINCT chs.date))::int
            ELSE NULL END
        FROM channel_historical_stats chs
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '7 days'
            AND chs.data->>'interactions' IS NOT NULL
    ) AS avg_interactions_7d,
    (
        SELECT CASE WHEN COUNT(DISTINCT chs.date) >= 7
            THEN (SUM((chs.data->>'interactions')::bigint) / COUNT(DISTINCT chs.date))::int
            ELSE NULL END
        FROM channel_historical_stats chs
        WHERE chs.channel_id = c.id
            AND chs.date >= CURRENT_DATE - INTERVAL '30 days'
            AND chs.data->>'interactions' IS NOT NULL
    ) AS avg_interactions_30d,
    (
        SELECT CASE WHEN total_views > 0
            THEN total_interactions::float / total_views
            ELSE NULL END
        FROM (
            SELECT
                SUM((chs.data->>'interactions')::bigint) AS total_interactions,
                SUM(vbs.val::bigint) AS total_views
            FROM channel_historical_stats chs,
                jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
            WHERE chs.channel_id = c.id
                AND chs.date >= CURRENT_DATE - INTERVAL '7 days'
                AND chs.data->>'interactions' IS NOT NULL
        ) sub
        WHERE (
            SELECT COUNT(DISTINCT chs2.date)
            FROM channel_historical_stats chs2
            WHERE chs2.channel_id = c.id
                AND chs2.date >= CURRENT_DATE - INTERVAL '7 days'
        ) >= 7
    ) AS engagement_rate_7d,
    (
        SELECT CASE WHEN total_views > 0
            THEN total_interactions::float / total_views
            ELSE NULL END
        FROM (
            SELECT
                SUM((chs.data->>'interactions')::bigint) AS total_interactions,
                SUM(vbs.val::bigint) AS total_views
            FROM channel_historical_stats chs,
                jsonb_each_text(chs.data->'views_by_source') AS vbs(key, val)
            WHERE chs.channel_id = c.id
                AND chs.date >= CURRENT_DATE - INTERVAL '30 days'
                AND chs.data->>'interactions' IS NOT NULL
        ) sub
        WHERE (
            SELECT COUNT(DISTINCT chs2.date)
            FROM channel_historical_stats chs2
            WHERE chs2.channel_id = c.id
                AND chs2.date >= CURRENT_DATE - INTERVAL '30 days'
        ) >= 7
    ) AS engagement_rate_30d
FROM channels c
LEFT JOIN channel_info ci ON ci.channel_id = c.id
WHERE c.deleted_at IS NULL AND c.is_listed = true;

CREATE UNIQUE INDEX idx_channel_marketplace_channel_id ON channel_marketplace(channel_id);
CREATE INDEX idx_channel_marketplace_subscribers ON channel_marketplace(subscribers DESC NULLS LAST);
CREATE INDEX idx_channel_marketplace_avg_views_7d ON channel_marketplace(avg_daily_views_7d DESC NULLS LAST);
