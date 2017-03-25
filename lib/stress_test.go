package pewpew

import (
	"reflect"
	"testing"
)

func TestValidateTargets(t *testing.T) {
	cases := []struct {
		s      StressConfig
		hasErr bool
	}{
		{StressConfig{}, true}, //multiple things uninitialized
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       0,
					Concurrency: DefaultConcurrency,
					Timeout:     DefaultTimeout,
					Method:      DefaultMethod,
				},
			},
		}, true}, //zero count
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       DefaultCount,
					Concurrency: 0,
					Timeout:     DefaultTimeout,
					Method:      DefaultMethod,
				},
			},
		}, true}, //zero concurrency
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       10,
					Concurrency: 20,
					Timeout:     DefaultTimeout,
					Method:      DefaultMethod,
				},
			},
		}, true}, //concurrency > count
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       DefaultCount,
					Concurrency: DefaultConcurrency,
					Timeout:     DefaultTimeout,
					Method:      "",
				},
			},
		}, true}, //empty method
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       DefaultCount,
					Concurrency: DefaultConcurrency,
					Timeout:     "",
					Method:      DefaultMethod,
				},
			},
		}, false}, //empty timeout string okay
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       DefaultCount,
					Concurrency: DefaultConcurrency,
					Timeout:     "unparseable",
					Method:      DefaultMethod,
				},
			},
		}, true}, //invalid time string
		{StressConfig{
			Targets: []Target{
				{
					URL:         DefaultURL,
					Count:       DefaultCount,
					Concurrency: DefaultConcurrency,
					Timeout:     "1ms",
					Method:      DefaultMethod,
				},
			},
		}, true}, //timeout too short

		//good cases
		{*NewStressConfig(), false},
	}
	for _, c := range cases {
		err := validateTargets(c.s)
		if (err != nil) != c.hasErr {
			t.Errorf("validateTargets(%+v) err: %t wanted %t", c.s, (err != nil), c.hasErr)
		}
	}
}

func TestBuildRequest(t *testing.T) {
	cases := []struct {
		target Target
		hasErr bool
	}{
		{Target{}, true},                          //empty url
		{Target{URL: ""}, true},                   //empty url
		{Target{URL: "", RegexURL: true}, true},   //empty url
		{Target{URL: "(*", RegexURL: true}, true}, //invalid regex
		{Target{URL: "asdf"}, true},               //invalid hostname
		{Target{URL: "localhost"}, true},          //missing scheme
		{Target{URL: "http://"}, true},            //empty hostname
		{Target{URL: "http://localhost",
			BodyFilename: "/thisfiledoesnotexist"}, true}, //bad file
		{Target{URL: "http://localhost",
			Headers: ",,,"}, true}, //invalid headers
		{Target{URL: "http://localhost",
			Headers: "a:b,c,d"}, true}, //invalid headers
		{Target{URL: "http://localhost",
			Cookies: ";;;"}, true}, //invalid cookies
		{Target{URL: "http://localhost",
			Cookies: "a=b;c;d"}, true}, //invalid cookies
		{Target{URL: "http://localhost",
			BasicAuth: "user:"}, true}, //invalid basic auth
		{Target{URL: "http://localhost",
			BasicAuth: ":pass"}, true}, //invalid basic auth
		{Target{URL: "http://localhost",
			BasicAuth: "::"}, true}, //invalid basic auth

		//good cases
		{Target{URL: "http://localhost:80"}, false},
		{Target{URL: "http://localhost",
			Method: "POST",
			Body:   "data"}, false},
		{Target{URL: "https://www.github.com"}, false},
		{Target{URL: "http://github.com"}, false},
		{Target{URL: "http://localhost:80/path/?param=val&another=one",
			Headers:   "Accept-Encoding:gzip, Content-Type:application/json",
			Cookies:   "a=b;c=d",
			UserAgent: "pewpewpew",
			BasicAuth: "user:pass"}, false},
	}
	for _, c := range cases {
		// req, err := buildRequest(c.target)
		_, err := buildRequest(c.target)
		if (err != nil) != c.hasErr {
			t.Errorf("buildRequest(%+v) err: %t wanted: %t", c.target, (err != nil), c.hasErr)
		}
	}
}

func TestParseKeyValString(t *testing.T) {
	cases := []struct {
		str    string
		delim1 string
		delim2 string
		want   map[string]string
		hasErr bool
	}{
		{"", "", "", map[string]string{}, true},
		{"", ":", ";", map[string]string{}, true},
		{"", ":", ":", map[string]string{}, true},
		{"abc:123;", ";", ":", map[string]string{"abc": "123"}, true},
		{"abc:123", ";", ":", map[string]string{"abc": "123"}, false},
		{"key1: val2, key3 : val4,key5:val6", ",", ":", map[string]string{"key1": "val2", "key3": "val4", "key5": "val6"}, false},
	}
	for _, c := range cases {
		result, err := parseKeyValString(c.str, c.delim1, c.delim2)
		if (err != nil) != c.hasErr {
			t.Errorf("parseKeyValString(%q, %q, %q) err: %t wanted %t", c.str, c.delim1, c.delim2, (err != nil), c.hasErr)
			continue
		}
		if !reflect.DeepEqual(result, c.want) {
			t.Errorf("parseKeyValString(%q, %q, %q) == %v wanted %v", c.str, c.delim1, c.delim2, result, c.want)
		}
	}
}
