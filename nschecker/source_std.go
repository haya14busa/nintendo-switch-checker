//+build !appengine

package nschecker

func init() {
	Sources = append(Sources,
		Source{
			Name:        "Yodobashi - Nintendo Switch Joy-Con(L)/(R)グレー [Nintendo Switch本体]",
			URL:         "http://www.yodobashi.com/product/100000001003431565/",
			SoldOutText: `<div class="salesInfo"><p>予定数の販売を終了しました</p></div>`,
		},
		Source{
			Name:        "Yodobashi - Nintendo Switch Joy-Con(L)ネオンブルー/(R)ネオンレッド [Nintendo Switch本体]",
			URL:         "http://www.yodobashi.com/product/100000001003431566/",
			SoldOutText: `<div class="salesInfo"><p>予定数の販売を終了しました</p></div>`,
		},
		Source{
			Name:        "Joshin - Nintendo Switch 本体【Joy-Con(L)/(R) グレー】",
			URL:         "http://joshinweb.jp/game/40519/4902370535709.html",
			SoldOutText: `<span class="fsL"><font color="blue"><b>販売休止中です</b></font><br></span>`,
		},
		Source{
			Name:        "Joshin - Nintendo Switch 本体【Joy-Con(L) ネオンブルー/(R) ネオンレッド】",
			URL:         "http://joshinweb.jp/game/40519/4902370535716.html",
			SoldOutText: `<span class="fsL"><font color="blue"><b>販売休止中です</b></font><br></span>`,
		},
	)
}
