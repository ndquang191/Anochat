"use client";

export type Language = "vi" | "en";

export const LANGUAGE_STORAGE_KEY = "language";

const translations = {
	vi: {
		accountSettings: "Cài đặt tài khoản",
		accountSettingsDescription: "Thay đổi thông tin cá nhân của bạn tại đây.",
		accountSuspended: "Tài khoản đã bị khóa",
		adminFailedToAddWord: "Không thể thêm từ cấm",
		adminFailedToBanUser: "Không thể chặn người dùng",
		adminFailedToRemoveWord: "Không thể xóa từ cấm",
		adminFailedToUnbanUser: "Không thể bỏ chặn người dùng",
		adminFailedToUpdateWord: "Không thể cập nhật từ cấm",
		adminUserBanned: "Đã chặn người dùng",
		adminUserUnbanned: "Đã bỏ chặn người dùng",
		adminWordAdded: "Đã thêm từ cấm",
		adminWordUpdated: "Đã cập nhật từ cấm",
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
		joinQueueShortcut: "Tham gia hàng chờ (Ctrl+Enter)",
		joinQueueSuccess: "Đã tham gia hàng chờ thành công!",
		leaveChatRoom: "Rời phòng chat",
		leaveChatRoomShortcut: "Rời phòng chat (Ctrl+Enter)",
		leaveChatRoomConfirmDescription: "Bạn có chắc muốn rời phòng chat này không?",
		leaveChatRoomSuccess: "Đã rời phòng chat",
		leaveQueue: "Nhấn để rời hàng chờ",
		leaveQueueShortcut: "Rời hàng chờ (Ctrl+Enter)",
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
		partnerLeftInlineNotice: "Người dùng đã rời phòng chat",
		pleaseTryAgain: "Vui lòng thử lại",
		pleaseWait: "Vui lòng chờ trong giây lát",
		privateProfile: "Riêng tư",
		profileUpdated: "Cập nhật thông tin thành công!",
		publicProfile: "Công khai thông tin",
		queueing: "Đang trong hàng chờ...",
		queueingDescription: "Vui lòng chờ trong khi chúng tôi tìm kiếm người chat cho bạn",
		report: "Báo cáo",
		reportDialogDescription: "Bạn có chắc muốn báo cáo người dùng này không?",
		reportDialogTitle: "Báo cáo người dùng",
		reportConfirmDescription: "Bạn có chắc muốn báo cáo người dùng này không?",
		reportConfirmTitle: "Báo cáo người dùng",
		reportSubmitted: "Đã gửi báo cáo",
		reportUser: "Báo cáo người dùng",
		saveChanges: "Lưu thay đổi",
		somethingWentWrong: "Có lỗi xảy ra",
		saving: "Đang lưu...",
		send: "Gửi",
		settings: "Cài đặt",
		sidebarToggleShortcut: "Mở/đóng sidebar (Ctrl+B)",
		theme: "Màu giao diện",
		turnOffSound: "Tắt âm thanh",
		turnOnSound: "Bật âm thanh",
		user: "Người dùng",
		userInfoSaveSuccess: "Thông tin đã được lưu thành công!",
		vietnamese: "Tiếng Việt",
		yearsOld: "{age} tuổi",
		yourMessagePlaceholder: "Nhập tin nhắn của bạn...",
	},
	en: {
		accountSettings: "Account settings",
		accountSettingsDescription: "Update your personal information here.",
		accountSuspended: "Your account has been suspended",
		adminFailedToAddWord: "Failed to add banned word",
		adminFailedToBanUser: "Failed to ban user",
		adminFailedToRemoveWord: "Failed to remove banned word",
		adminFailedToUnbanUser: "Failed to unban user",
		adminFailedToUpdateWord: "Failed to update banned word",
		adminUserBanned: "User banned",
		adminUserUnbanned: "User unbanned",
		adminWordAdded: "Banned word added",
		adminWordUpdated: "Banned word updated",
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
		joinQueueShortcut: "Join queue (Ctrl+Enter)",
		joinQueueSuccess: "Joined the queue successfully!",
		leaveChatRoom: "Leave chat room",
		leaveChatRoomShortcut: "Leave chat room (Ctrl+Enter)",
		leaveChatRoomConfirmDescription: "Are you sure you want to leave this chat room?",
		leaveChatRoomSuccess: "Left the chat room",
		leaveQueue: "Click to leave the queue",
		leaveQueueShortcut: "Leave queue (Ctrl+Enter)",
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
		partnerLeftInlineNotice: "The other user left the chat",
		pleaseTryAgain: "Please try again",
		pleaseWait: "Please wait a moment",
		privateProfile: "Private",
		profileUpdated: "Profile updated successfully!",
		publicProfile: "Public profile",
		queueing: "Waiting in queue...",
		queueingDescription: "Please wait while we find someone for you to chat with",
		report: "Report",
		reportDialogDescription: "Are you sure you want to report this user?",
		reportDialogTitle: "Report user",
		reportConfirmDescription: "Are you sure you want to report this user?",
		reportConfirmTitle: "Report user",
		reportSubmitted: "Report submitted",
		reportUser: "Report user",
		saveChanges: "Save changes",
		somethingWentWrong: "Something went wrong",
		saving: "Saving...",
		send: "Send",
		settings: "Settings",
		sidebarToggleShortcut: "Toggle sidebar (Ctrl+B)",
		theme: "Theme color",
		turnOffSound: "Turn off sound",
		turnOnSound: "Turn on sound",
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
) : string {
	const template = translations[language][key] ?? translations.vi[key];

	if (!vars) {
		return String(template);
	}

	return Object.entries(vars).reduce(
		(result, [name, value]) => result.replaceAll(`{${name}}`, String(value)),
		String(template)
	);
}

export function translateStored(
	key: TranslationKey,
	vars?: Record<string, string | number>
) : string {
	return translate(getStoredLanguage(), key, vars);
}
