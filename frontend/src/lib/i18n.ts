"use client";

export type Language = "vi" | "en";

export const LANGUAGE_STORAGE_KEY = "language";

const translations = {
	vi: {
		accountSettings: "Cài đặt tài khoản",
		accountSettingsDescription: "Thay đổi thông tin cá nhân của bạn tại đây.",
		admin: "Admin",
		adminPanel: "Bảng quản trị",
		age: "Tuổi",
		agePlaceholder: "Nhập tuổi",
		anonymous: "Ẩn danh",
		appearance: "Giao diện",
		cancel: "Hủy",
		closeAdminPanel: "Đóng bảng quản trị",
		connectingWebSocket: "Đang kết nối WebSocket...",
		darkTheme: "Tối",
		english: "English",
		female: "Nữ",
		findPartnerDescription: "Vui lòng tham gia hàng chờ để tìm đối tác chat",
		gender: "Giới tính",
		interfaceLanguage: "Ngôn ngữ giao diện",
		joinQueue: "Tham gia hàng chờ",
		joinQueueSuccess: "Đã tham gia hàng chờ thành công!",
		leaveChatRoom: "Rời phòng chat",
		leaveChatRoomSuccess: "Đã rời phòng chat",
		leaveQueue: "Nhấn để rời hàng chờ",
		leaveQueueSuccess: "Đã rời khỏi hàng chờ!",
		loading: "Đang tải...",
		loadingUser: "Đang tải thông tin người dùng...",
		logout: "Đăng xuất",
		logoutSuccess: "Đăng xuất thành công!",
		male: "Nam",
		matchFound: "Đã tìm thấy đối tác chat!",
		matchFoundDescription: "Bạn đã được kết nối với người dùng khác",
		noChatRoom: "Chưa có phòng chat nào",
		other: "Khác",
		partnerLeft: "Đối tác đã rời phòng",
		partnerLeftDescription: "Bạn có thể tìm kiếm đối tác mới",
		pleaseTryAgain: "Vui lòng thử lại",
		pleaseWait: "Vui lòng chờ trong giây lát",
		privateProfile: "Riêng tư",
		profileUpdated: "Cập nhật thông tin thành công!",
		publicProfile: "Công khai thông tin",
		queueing: "Đang trong hàng chờ...",
		queueingDescription: "Vui lòng chờ trong khi chúng tôi tìm kiếm người chat cho bạn",
		report: "Báo cáo",
		reportConfirmDescription: "Bạn có chắc muốn báo cáo người dùng này không?",
		reportConfirmTitle: "Báo cáo người dùng",
		reportSubmitted: "Đã gửi báo cáo",
		reportUser: "Báo cáo người dùng",
		saveChanges: "Lưu thay đổi",
		saving: "Đang lưu...",
		send: "Gửi",
		settings: "Cài đặt",
		theme: "Màu giao diện",
		user: "Người dùng",
		userInfoSaveSuccess: "Thông tin đã được lưu thành công!",
		vietnamese: "Tiếng Việt",
		yearsOld: "{age} tuổi",
		yourMessagePlaceholder: "Nhập tin nhắn của bạn...",
	},
	en: {
		accountSettings: "Account settings",
		accountSettingsDescription: "Update your personal information here.",
		admin: "Admin",
		adminPanel: "Admin Panel",
		age: "Age",
		agePlaceholder: "Enter age",
		anonymous: "Anonymous",
		appearance: "Appearance",
		cancel: "Cancel",
		closeAdminPanel: "Close admin panel",
		connectingWebSocket: "Connecting to WebSocket...",
		darkTheme: "Dark",
		english: "English",
		female: "Female",
		findPartnerDescription: "Join the queue to find a chat partner",
		gender: "Gender",
		interfaceLanguage: "Interface language",
		joinQueue: "Join queue",
		joinQueueSuccess: "Joined the queue successfully!",
		leaveChatRoom: "Leave chat room",
		leaveChatRoomSuccess: "Left the chat room",
		leaveQueue: "Click to leave the queue",
		leaveQueueSuccess: "Left the queue!",
		loading: "Loading...",
		loadingUser: "Loading user information...",
		logout: "Log out",
		logoutSuccess: "Logged out successfully!",
		male: "Male",
		matchFound: "Chat partner found!",
		matchFoundDescription: "You have been connected with another user",
		noChatRoom: "No active chat room",
		other: "Other",
		partnerLeft: "Your partner left the room",
		partnerLeftDescription: "You can look for a new partner now",
		pleaseTryAgain: "Please try again",
		pleaseWait: "Please wait a moment",
		privateProfile: "Private",
		profileUpdated: "Profile updated successfully!",
		publicProfile: "Public profile",
		queueing: "Waiting in queue...",
		queueingDescription: "Please wait while we find someone for you to chat with",
		report: "Report",
		reportConfirmDescription: "Are you sure you want to report this user?",
		reportConfirmTitle: "Report user",
		reportSubmitted: "Report submitted",
		reportUser: "Report user",
		saveChanges: "Save changes",
		saving: "Saving...",
		send: "Send",
		settings: "Settings",
		theme: "Theme color",
		user: "User",
		userInfoSaveSuccess: "Your information has been saved successfully!",
		vietnamese: "Tiếng Việt",
		yearsOld: "{age} years old",
		yourMessagePlaceholder: "Type your message...",
	},
} as const;

export type TranslationKey = keyof typeof translations.vi;

export function isLanguage(value: string | null | undefined): value is Language {
	return value === "vi" || value === "en";
}

export function getStoredLanguage(): Language {
	if (typeof window === "undefined") {
		return "vi";
	}

	try {
		const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
		return isLanguage(stored) ? stored : "vi";
	} catch {
		return "vi";
	}
}

export function translate(
	language: Language,
	key: TranslationKey,
	vars?: Record<string, string | number>
) {
	const template = translations[language][key] ?? translations.vi[key];

	if (!vars) {
		return template;
	}

	return Object.entries(vars).reduce(
		(result, [name, value]) => result.replaceAll(`{${name}}`, String(value)),
		template
	);
}

export function translateStored(
	key: TranslationKey,
	vars?: Record<string, string | number>
) {
	return translate(getStoredLanguage(), key, vars);
}
