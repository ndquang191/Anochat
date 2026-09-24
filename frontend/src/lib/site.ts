const DEFAULT_SITE_URL = "http://localhost:3000";

export const SITE_NAME = "AnoChat";
export const SITE_TAGLINE = "Chat ẩn danh ngẫu nhiên tại Việt Nam";
export const SITE_SOCIAL_TITLE = "AnoChat — Nói điều thật lòng";
export const SITE_DESCRIPTION =
	"Gặp một người lạ và bắt đầu cuộc trò chuyện riêng tư — không profile, không feed, không áp lực.";
export const SITE_KEYWORDS = [
	"AnoChat",
	"anonymous chat VN",
	"chat ẩn danh",
	"chat ẩn danh Việt Nam",
	"nói chuyện với người lạ",
	"kết nối ngẫu nhiên",
	"random chat Việt Nam",
	"trò chuyện ẩn danh",
	"chat với người lạ",
];

function normalizeSiteUrl(url?: string) {
	if (!url) {
		return DEFAULT_SITE_URL;
	}

	try {
		return new URL(url).origin;
	} catch {
		return DEFAULT_SITE_URL;
	}
}

export function getSiteUrl() {
	return normalizeSiteUrl(process.env.NEXT_PUBLIC_SITE_URL?.trim());
}

export function getMetadataBase() {
	return new URL(getSiteUrl());
}

export function absoluteUrl(path = "/") {
	return new URL(path, getSiteUrl()).toString();
}
