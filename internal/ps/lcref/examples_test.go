package lcref

import (
	"regexp"
	"strings"
	"testing"
)

const sampleContent = `<p>Given an array...</p><pre><strong>Input:</strong> nums = [2,7,11,15], target = 9
<strong>Output:</strong> [0,1]
<strong>Explanation:</strong> nums[0] + nums[1] == 9.</pre><pre><strong>Input:</strong> nums = [3,2,4], target = 6
<strong>Output:</strong> [1,2]</pre><pre>그 외 코드 블록</pre>`

func TestParseExamples(t *testing.T) {
	got := ParseExamples(sampleContent)
	if len(got) != 2 {
		t.Fatalf("예제 2개여야 함: %+v", got)
	}
	if got[0].In != "nums = [2,7,11,15], target = 9" {
		t.Errorf("in[0]=%q", got[0].In)
	}
	if got[0].Out != "[0,1]" {
		t.Errorf("out[0]=%q", got[0].Out)
	}
	if got[1].Out != "[1,2]" {
		t.Errorf("out[1]=%q", got[1].Out)
	}
}

func TestHTMLToText(t *testing.T) {
	got := HTMLToText("<p>a&nbsp;&lt;b&gt;</p><ul><li>x</li></ul>")
	if !strings.Contains(got, "a <b>") || !strings.Contains(got, "• x") {
		t.Errorf("변환 결과 이상: %q", got)
	}
}

func TestHTMLToMarkdown(t *testing.T) {
	src := `<p>Given an <strong>array</strong> of <code>nums</code>.</p>` +
		`<pre><strong>Input:</strong> nums = [2,7]
<strong>Output:</strong> [0,1]</pre>` +
		`<ul><li><code>2 &lt;= nums.length &lt;= 10<sup>4</sup></code></li><li>follow up</li></ul>` +
		`<p><a href="https://x.com">link</a> and <img src="https://x.com/f.png"></p>`
	got := HTMLToMarkdown(src)
	for _, want := range []string{
		"**array**", "`nums`", "```\nInput: nums = [2,7]\nOutput: [0,1]\n```",
		"- `2 <= nums.length <= 10^4`", "- follow up", "[link](https://x.com)", "![img](https://x.com/f.png)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("마크다운 변환에서 %q 누락:\n%s", want, got)
		}
	}
	if regexp.MustCompile(`<\\s*/?[a-zA-Z]`).MatchString(got) || strings.Contains(got, "&lt;") {
		t.Errorf("태그/엔티티 잔여: %q", got)
	}
}
