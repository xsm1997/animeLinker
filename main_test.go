package main

import (
	"testing"
)

func TestProbeVideoName(t *testing.T) {
	in := []string{
		`[Airota&VCB-Studio] Koutetsujou no Kabaneri [Ma10p_1080p]`,
		`[Beatrice-Raws] Re Zero kara Hajimeru Isekai Seikatsu - The Frozen Bond [BDRip 1920x1080 HEVC FLAC]`,
		`[Snow-Raws] BANANA FISH`,
		`[Snow-Raws] 牙狼〈GARO〉-VANISING LINE-`,
		`[Snow-Raws] ソードアート・オンライン アリシゼーション War of Underworld 第01話 (BD 1920x1080 HEVC-YUV420P10 FLACx2)`,
		`[MakariHoshiyume&VCB-Studio] DanMachi [01][Ma10p_1080p][x265_2flac].sc.ass`,
		`DanMachi.sc.ass`,
		`DanMachi.mkv`,
		`[Snow-Raws] アルスラーン戦記 風塵乱舞 第01話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mp4`,
		`[Snow-Raws] Charlotte 第01話(BD 1920x1080 HEVC-YUV420P10 FLACx2).mp4`,
		`[EMD]Arslan Senki[12][GB_BIG5][X264_AAC][1280X720][7BAA2B61].mp4`,
		`[EMD]Arslan Senki[13.5][GB_BIG5][X264_AAC][1280X720][7BAA2B61].mp4`,
	}

	out1 := []string{
		`Koutetsujou no Kabaneri`,
		`Re Zero kara Hajimeru Isekai Seikatsu The Frozen Bond`,
		`BANANA FISH`,
		`牙狼〈GARO〉-VANISING LINE-`,
		`ソードアート・オンライン アリシゼーション War of Underworld 第01話`,
		`DanMachi [01].sc.ass`,
		`DanMachi.sc.ass`,
		`DanMachi.mkv`,
		`アルスラーン戦記 風塵乱舞 第01話.mp4`,
		`Charlotte 第01話.mp4`,
		`Arslan Senki[12].mp4`,
		`Arslan Senki[13.5].mp4`,
	}

	for i, data := range in {
		o1 := out1[i]

		r1 := probeVideoName(data)

		if o1 != "" {
			if r1 != o1 {
				t.Errorf("Data %s: excepted %s, got %s", data, o1, r1)
			}
		}
	}
}

func TestGetEpisode(t *testing.T) {
	in := []string{
		`ソードアート・オンライン アリシゼーション War of Underworld 第01話`,
		`DanMachi [01].sc.ass`,
		`アルスラーン戦記 風塵乱舞 第01話.mp4`,
		`Charlotte 第01話.mp4`,
		`Arslan Senki[12].mp4`,
		`Arslan Senki[13.5].mp4`,
	}

	out1 := []string{
		`01`,
		`01`,
		`01`,
		`01`,
		`12`,
		`13.5`,
	}

	for i, data := range in {
		o1 := out1[i]

		r1 := getEpisode(data)

		if o1 != "" {
			if r1 != o1 {
				t.Errorf("Data %s: excepted %s, got %s", data, o1, r1)
			}
		}
	}
}
