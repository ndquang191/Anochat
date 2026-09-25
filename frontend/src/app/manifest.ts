import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
	return {
		id: "/",
		name: "AnoChat",
		short_name: "AnoChat",
		description: "Trò chuyện ẩn danh và kết nối với một người mới.",
		start_url: "/",
		scope: "/",
		display: "standalone",
		background_color: "#f7fafc",
		theme_color: "#516b91",
		categories: ["social"],
		icons: [
			{ src: "/icons/icon-192.png", sizes: "192x192", type: "image/png", purpose: "any" },
			{ src: "/icons/icon-512.png", sizes: "512x512", type: "image/png", purpose: "any" },
			{ src: "/icons/icon-maskable.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
		],
	};
}
