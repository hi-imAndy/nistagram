package service

import (
	"encoding/json"
	"net/http"
	"nistagram/agent/dto"
	"nistagram/agent/model"
	"nistagram/agent/util"
	"strings"
)

type CampaignService struct {

}

func (service *CampaignService) SaveCampaignReport(campaignId uint) error{
	resp, err := util.NistagramRequest(http.MethodGet, "/agent-api/statistics/" + util.Uint2String(campaignId),
		nil, map[string]string{})

	if err != nil{
		return err
	}

	var stat dto.StatisticsDTO

	err = json.NewDecoder(resp.Body).Decode(&stat)
	if err != nil{
		return err
	}

	var report model.CampaignReport
	var basicInfo model.BasicInformation
	var overallStat model.OverallStatistics
	var oStats model.Stats
	var paramStat []model.ParametersStatistics

	basicInfo.CampaignId = stat.CampaignId
	basicInfo.PostID = stat.Campaign.PostID
	basicInfo.AgentID = stat.Campaign.AgentID
	basicInfo.CampaignType = stat.Campaign.CampaignType
	basicInfo.Start = stat.Campaign.Start
	//TODO: at the end - basicInfo.End

	//overall stats
	for _, event := range stat.Events{
		switch strings.ToLower(event.EventType){
		case "like":
			oStats.Likes += 1
		case "dislike":
			oStats.Dislikes += 1
		case "like_reset":
			oStats.LikeResets += 1
		case "dislike_reset":
			oStats.DislikeResets += 1
		case "comment":
			oStats.Comments += 1
		case "visit":
			oStats.TotalSiteVisits += 1
			oStats.AddSpecificSite(event.WebSite)
		}
	}

	//params stats
	for _, params := range stat.Campaign.CampaignParameters{
		var ps model.ParametersStatistics
		ps.Start = params.Start
		ps.End = params.End
		ps.Timestamps = params.Timestamps

		var infNotAccepted []string
		for _, req := range params.CampaignRequests{
			if strings.ToLower(req.RequestStatus) == "declined"{
				infNotAccepted = append(infNotAccepted, req.InfluencerUsername)
			}
		}

		for _, event := range stat.Events{
			if event.Timestamp.After(params.Start) && event.Timestamp.Before(params.End){
				if event.InfluencerId != 0{
					ps.AddEventForInf(event.InfluencerUsername, event.EventType, event.WebSite)
				}else if len(event.Interests) != 0{
					ps.AddEventForInterest(event.Interests, event.EventType, event.WebSite)
				}else{
					//TODO: direct event
				}
			}
		}

		ps.InfluencerWhoDidNotAccept = infNotAccepted
		paramStat = append(paramStat, ps)
	}

	overallStat.Stats = oStats
	report.BasicInformation = basicInfo
	report.OverallStatistics = overallStat
	report.ParametersStatistics = paramStat

	return nil
}