package blocker

func ReverseHostName(host string) string {
	last := len(host)
	if last == 0 {
		return host
	}

	app := 0
	ret := make([]byte, len(host))
	for i := last - 1; i >= 0; i-- {
		if i == 0 {
			// append whatever's left
			copy(ret[app:app+last-i], host[i:last])
			break
		}

		if host[i] != '.' {
			continue
		}

		copy(ret[app:app+last-i-1], host[i+1:last])
		ret[app+last-i-1] = '.'

		app += last - i
		last = i
	}

	return string(ret)
}
