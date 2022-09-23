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
		`[MakariHoshiyume&VCB-Studio] DanMachi [Ma10p_1080p][x265_2flac]`,
		`DanMachi.sc.ass`,
		`DanMachi.mkv`,
		`[Snow-Raws] アルスラーン戦記 風塵乱舞 (BD 1920x1080 HEVC-YUV420P10 FLAC)`,
		`[Snow-Raws] Charlotte(BD 1920x1080 HEVC-YUV420P10 FLACx2)`,
		`[EMD]Arslan Senki[GB_BIG5][X264_AAC][1280X720][7BAA2B61]`,
		`[EMD]Arslan Senki[GB_BIG5][X264_AAC][1280X720][7BAA2B61]`,
		`[2020][Dungeon ni Deai o Motomeru no wa Machigatte Iru Darouka III][BDRIP][1080P][1-12Fin+SP]`,
		`[AI-Raws][牙狼_GARO Animation Series][BDRip][MKV]`,
	}

	out1 := []string{
		`Koutetsujou no Kabaneri`,
		`Re Zero kara Hajimeru Isekai Seikatsu The Frozen Bond`,
		`BANANA FISH`,
		`牙狼〈GARO〉-VANISING LINE-`,
		`ソードアート・オンライン アリシゼーション War of Underworld`,
		`DanMachi`,
		`DanMachi.sc.ass`,
		`DanMachi.mkv`,
		`アルスラーン戦記 風塵乱舞`,
		`Charlotte`,
		`Arslan Senki`,
		`Arslan Senki`,
		`Dungeon ni Deai o Motomeru no wa Machigatte Iru Darouka III`,
		`牙狼_GARO Animation Series`,
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
		`2021 abaaba 2022 [12].mp4`,
		`[Airota&VCB-Studio] Koutetsujou no Kabaneri - 10 [Ma10p_1080p]`,
		`[Beatrice-Raws] Re Zero kara Hajimeru Isekai Seikatsu - The Frozen Bond - 10.5 [BDRip 1920x1080 HEVC FLAC]`,
		`[Snow-Raws] BANANA FISH [01]`,
		`[Snow-Raws] 牙狼〈GARO〉-VANISING LINE- [01][BDRip 1920x1080 HEVC FLAC]`,
		`[Snow-Raws] ソードアート・オンライン アリシゼーション War of Underworld 第01話 (BD 1920x1080 HEVC-YUV420P10 FLACx2)`,
		`[MakariHoshiyume&VCB-Studio] DanMachi [01][Ma10p_1080p][x265_2flac].sc.ass`,
		`DanMachi.sc.ass`,
		`DanMachi.mkv`,
		`[DanMachi S3][02][BDRIP][1080P][H264_FLAC].mkv`,
		`DanMachi 10.mkv`,
		`Charlotte [第01話].mp4`,
		`世界最高の暗殺者、異世界貴族に転生する メニュー動画 vol1 (BD 1920x1080 x265 ALAC).mp4`,
		`[ANK-Raws] 血界戦線 CM01 (BDrip 1920x1080 HEVC-YUV420P10 FLAC).mkv`,
		`[AI-Raws][アニメ BD] 牙狼-GARO- -炎の刻印- ゆるがろ #01 (H264 10bit 1920x1080 FLAC)[1B793118].mkv`,
	}

	out1 := []string{
		`01`,
		`01`,
		`01`,
		`01`,
		`12`,
		`13.5`,
		`12`,
		`10`,
		`10.5`,
		`01`,
		`01`,
		`01`,
		`01`,
		``,
		``,
		`02`,
		`10`,
		`01`,
		``,
		`CM01`,
		`01`,
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
