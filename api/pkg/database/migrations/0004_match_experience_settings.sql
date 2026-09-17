ALTER TABLE match_settings
ADD COLUMN IF NOT EXISTS rematch_cooldown_seconds INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS queue_display_mode VARCHAR(16) NOT NULL DEFAULT 'message',
ADD COLUMN IF NOT EXISTS queue_count_minimum INTEGER NOT NULL DEFAULT 5,
ADD COLUMN IF NOT EXISTS queue_message_vi TEXT NOT NULL DEFAULT 'Anochat còn mới nên đôi khi việc tìm người trò chuyện sẽ mất một chút thời gian. Cảm ơn bạn đã chờ nhé!',
ADD COLUMN IF NOT EXISTS queue_message_en TEXT NOT NULL DEFAULT 'Anochat is still new, so finding someone may take a little time. Thanks for waiting!';

ALTER TABLE match_settings
ADD CONSTRAINT match_settings_rematch_cooldown_check
CHECK (rematch_cooldown_seconds IN (0, 21600, 43200, 86400, 604800)),
ADD CONSTRAINT match_settings_queue_display_mode_check
CHECK (queue_display_mode IN ('hidden', 'message', 'count')),
ADD CONSTRAINT match_settings_queue_count_minimum_check
CHECK (queue_count_minimum BETWEEN 1 AND 1000);
