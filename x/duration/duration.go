package duration

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d Duration) As() time.Duration {
	return time.Duration(d)
}

// MarshalJSON 将 Duration 转换为字符串（如 "5s"）
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON 将字符串（如 "5s"）解析为 Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}
