package dto

import (
	"LinkTracker/internal/domain"
	scrapperdto "LinkTracker/internal/infrastructure/dto/dto_scrapper"
)

func LinkToLinkRequestDTO(link *domain.Link) scrapperdto.LinkRequest {
	if link == nil {
		return scrapperdto.LinkRequest{}
	}

	return scrapperdto.LinkRequest{Link: &link.URL, Filters: &link.Filters, Tags: &link.Tags}
}

func LinkToLinkResponseDTO(link *domain.Link) scrapperdto.LinkResponse {
	if link == nil {
		return scrapperdto.LinkResponse{}
	}

	return scrapperdto.LinkResponse{Url: &link.URL, Id: &link.ID, Tags: &link.Tags, Filters: &link.Filters}
}

func LinkRequestDTOToLink(linkRequest scrapperdto.LinkRequest) (domain.Link, error) {
	if linkRequest.Link == nil || *linkRequest.Link == "" {
		return domain.Link{}, domain.ErrNoRequiredAttribute{Attribute: "link"}
	}

	if linkRequest.Tags == nil {
		linkRequest.Tags = &[]string{}
	}

	if linkRequest.Filters == nil {
		linkRequest.Filters = &[]string{}
	}

	var link domain.Link
	link.URL = *linkRequest.Link
	link.Tags = *linkRequest.Tags
	link.Filters = *linkRequest.Filters

	return link, nil
}

func LinksToListLinksResponseDTO(links []domain.Link) scrapperdto.ListLinksResponse {
	linksResponse := make([]scrapperdto.LinkResponse, len(links))
	for i := range links {
		linksResponse[i] = scrapperdto.LinkResponse{
			Id:      &links[i].ID,
			Url:     &links[i].URL,
			Tags:    &links[i].Tags,
			Filters: &links[i].Filters,
		}
	}

	length := int32(len(linksResponse)) //nolint:gosec //api contract compliance(+ overflow is unlikely to be possible in real life)

	return scrapperdto.ListLinksResponse{Links: &linksResponse, Size: &length}
}

func RemoveLinkRequestDTOToLink(removeLinkRequestDTO scrapperdto.RemoveLinkRequest) (domain.Link, error) {
	if removeLinkRequestDTO.Link == nil || *removeLinkRequestDTO.Link == "" {
		return domain.Link{}, domain.ErrNoRequiredAttribute{Attribute: "link"}
	}

	return domain.Link{URL: *removeLinkRequestDTO.Link}, nil
}

func LinkToRemoveListRequestDTO(link *domain.Link) scrapperdto.RemoveLinkRequest {
	if link == nil {
		return scrapperdto.RemoveLinkRequest{}
	}

	return scrapperdto.RemoveLinkRequest{
		Link: &link.URL,
	}
}
