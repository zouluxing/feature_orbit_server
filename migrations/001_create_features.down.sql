DROP TRIGGER IF EXISTS trg_features_updated_at ON features;
DROP FUNCTION IF EXISTS set_features_updated_at();
DROP TABLE IF EXISTS features CASCADE;
