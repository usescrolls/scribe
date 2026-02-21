package scribe

import "time"

// IsUpdateNotificationsDisabled checks if update notifications are suppressed.
func IsUpdateNotificationsDisabled() (bool, error) {
	config, err := LoadConfig()
	if err != nil {
		return false, err
	}
	return config.UpdateNotificationsDisabled, nil
}

// SetUpdateNotificationsDisabled enables or disables update notifications.
func SetUpdateNotificationsDisabled(disabled bool) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.UpdateNotificationsDisabled = disabled
	return SaveConfig(config)
}

// GetLastUpdateCheck returns the timestamp of the last update check.
func GetLastUpdateCheck() (time.Time, error) {
	config, err := LoadConfig()
	if err != nil {
		return time.Time{}, err
	}
	if config.LastUpdateCheck == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, config.LastUpdateCheck)
}

// SetLastUpdateCheck records the current time as the last update check.
func SetLastUpdateCheck() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.LastUpdateCheck = time.Now().Format(time.RFC3339)
	return SaveConfig(config)
}

// ShouldCheckForUpdate returns true if enough time has elapsed since the last check.
func ShouldCheckForUpdate(interval time.Duration) bool {
	lastCheck, err := GetLastUpdateCheck()
	if err != nil {
		return true
	}
	if lastCheck.IsZero() {
		return true
	}
	return time.Since(lastCheck) > interval
}
