-- Seed: idac_area_metadata.area_code (ISO codes shown in the Discord area dropdown)
--
-- Countries use ISO 3166-1 alpha-3 (e.g. Vietnam -> VNM).
-- Japan prefectures use ISO 3166-2:JP subdivision codes (e.g. Tokyo -> JP-13).
-- Keyed by sega_area_code (stable). `ELSE area_code` preserves any value not listed
-- (e.g. "World" / area-all is intentionally left unset -> dropdown shows just "World").
--
-- Idempotent: safe to re-run.

ALTER TABLE idac_area_metadata ADD COLUMN IF NOT EXISTS area_code text;

UPDATE idac_area_metadata
SET area_code = CASE sega_area_code
    -- Countries (ISO 3166-1 alpha-3)
    WHEN 'area-47' THEN 'MAC' -- Macau
    WHEN 'area-48' THEN 'HKG' -- Hong Kong
    WHEN 'area-49' THEN 'KOR' -- Korea
    WHEN 'area-50' THEN 'MYS' -- Malaysia
    WHEN 'area-51' THEN 'SGP' -- Singapore
    WHEN 'area-52' THEN 'TWN' -- Taiwan
    WHEN 'area-53' THEN 'IDN' -- Indonesia
    WHEN 'area-54' THEN 'PHL' -- Philippines
    WHEN 'area-55' THEN 'THA' -- Thailand
    WHEN 'area-56' THEN 'USA' -- USA
    WHEN 'area-57' THEN 'VNM' -- Vietnam
    WHEN 'area-58' THEN 'MMR' -- Myanmar
    WHEN 'area-59' THEN 'AUS' -- Australia
    WHEN 'area-60' THEN 'NZL' -- New Zealand
    WHEN 'area-61' THEN 'KHM' -- Cambodia

    -- Japan prefectures (ISO 3166-2:JP)
    WHEN 'area-0'  THEN 'JP-01' -- Hokkaido
    WHEN 'area-1'  THEN 'JP-02' -- Aomori
    WHEN 'area-2'  THEN 'JP-03' -- Iwate
    WHEN 'area-3'  THEN 'JP-04' -- Miyagi
    WHEN 'area-6'  THEN 'JP-05' -- Akita
    WHEN 'area-5'  THEN 'JP-06' -- Yamagata
    WHEN 'area-4'  THEN 'JP-07' -- Fukushima
    WHEN 'area-7'  THEN 'JP-08' -- Ibaraki
    WHEN 'area-8'  THEN 'JP-09' -- Tochigi
    WHEN 'area-9'  THEN 'JP-10' -- Gunma
    WHEN 'area-11' THEN 'JP-11' -- Saitama
    WHEN 'area-10' THEN 'JP-12' -- Chiba
    WHEN 'area-12' THEN 'JP-13' -- Tokyo
    WHEN 'area-13' THEN 'JP-14' -- Kanagawa
    WHEN 'area-15' THEN 'JP-15' -- Niigata
    WHEN 'area-17' THEN 'JP-16' -- Toyama
    WHEN 'area-18' THEN 'JP-17' -- Ishikawa
    WHEN 'area-23' THEN 'JP-18' -- Fukui
    WHEN 'area-14' THEN 'JP-19' -- Yamanashi
    WHEN 'area-16' THEN 'JP-20' -- Nagano
    WHEN 'area-21' THEN 'JP-21' -- Gifu
    WHEN 'area-20' THEN 'JP-22' -- Shizuoka
    WHEN 'area-19' THEN 'JP-23' -- Aichi
    WHEN 'area-22' THEN 'JP-24' -- Mie
    WHEN 'area-27' THEN 'JP-25' -- Shiga
    WHEN 'area-25' THEN 'JP-26' -- Kyoto
    WHEN 'area-24' THEN 'JP-27' -- Osaka
    WHEN 'area-29' THEN 'JP-28' -- Hyogo
    WHEN 'area-26' THEN 'JP-29' -- Nara
    WHEN 'area-28' THEN 'JP-30' -- Wakayama
    WHEN 'area-31' THEN 'JP-31' -- Tottori
    WHEN 'area-32' THEN 'JP-32' -- Shimane
    WHEN 'area-33' THEN 'JP-33' -- Okayama
    WHEN 'area-30' THEN 'JP-34' -- Hiroshima
    WHEN 'area-34' THEN 'JP-35' -- Yamaguchi
    WHEN 'area-35' THEN 'JP-36' -- Tokushima
    WHEN 'area-36' THEN 'JP-37' -- Kagawa
    WHEN 'area-37' THEN 'JP-38' -- Ehime
    WHEN 'area-38' THEN 'JP-39' -- Kochi
    WHEN 'area-39' THEN 'JP-40' -- Fukuoka
    WHEN 'area-40' THEN 'JP-41' -- Saga
    WHEN 'area-41' THEN 'JP-42' -- Nagasaki
    WHEN 'area-42' THEN 'JP-43' -- Kumamoto
    WHEN 'area-43' THEN 'JP-44' -- Oita
    WHEN 'area-44' THEN 'JP-45' -- Miyazaki
    WHEN 'area-45' THEN 'JP-46' -- Kagoshima
    WHEN 'area-46' THEN 'JP-47' -- Okinawa

    ELSE area_code
END;
