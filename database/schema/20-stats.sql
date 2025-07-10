CREATE TABLE stats
(
    id          INT PRIMARY KEY DEFAULT 0 CHECK (id = 0),
    block_count BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert initial row
INSERT INTO stats (block_count) VALUES (0);

-- Function to update block count
CREATE OR REPLACE FUNCTION update_block_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE stats 
    SET block_count = CASE 
        WHEN block_count = 0 THEN (SELECT COUNT(*) FROM block)
        ELSE block_count + 1
    END,
    updated_at = NOW()
    WHERE id = 0;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update block count on insert
CREATE TRIGGER trigger_update_block_count
    AFTER INSERT ON block
    FOR EACH ROW
    EXECUTE FUNCTION update_block_count();
