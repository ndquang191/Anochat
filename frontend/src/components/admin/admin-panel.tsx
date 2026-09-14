"use client";

import { useEffect, useState } from "react";
import { BannedWordsTab } from "./banned-words-tab";
import { ReportsTab } from "./reports-tab";
import { BannedUsersTab } from "./banned-users-tab";
import { OverviewTab } from "./overview-tab";
import { useLanguage } from "@/contexts/theme";
import type { TranslationKey } from "@/lib/i18n";
import { secureStorage } from "@/lib/secure-storage";

type Tab = "overview" | "words" | "reports" | "banned";

const TABS: { id: Tab; labelKey: TranslationKey }[] = [
	{ id: "overview", labelKey: "adminOverview" },
	{ id: "reports", labelKey: "adminReports" },
	{ id: "banned", labelKey: "adminBannedUsers" },
	{ id: "words", labelKey: "adminBannedWords" },
];

function readLegacyTab(): Tab | null {
	try {
		return localStorage.getItem("admin_active_tab") as Tab | null;
	} catch {
		return null;
	}
}

function removeLegacyTab() {
	try {
		localStorage.removeItem("admin_active_tab");
	} catch {}
}

export function AdminPanel() {
	const { t } = useLanguage();
	const [activeTab, setActiveTab] = useState<Tab>("overview");

	useEffect(() => {
		let active = true;
		void secureStorage.get<Tab>("admin-active-tab").then(async (savedTab) => {
			let resolvedTab = savedTab;
			const legacyTab = readLegacyTab();
			if (
				!resolvedTab &&
				legacyTab &&
				TABS.some((tab) => tab.id === legacyTab)
			) {
				resolvedTab = legacyTab;
				if (await secureStorage.set("admin-active-tab", legacyTab)) {
					removeLegacyTab();
				}
			} else if (savedTab) {
				removeLegacyTab();
			}
			if (active && resolvedTab && TABS.some((tab) => tab.id === resolvedTab)) {
				setActiveTab(resolvedTab);
			}
		});
		return () => {
			active = false;
		};
	}, []);

	const handleTabChange = (tab: Tab) => {
		setActiveTab(tab);
		void secureStorage.set("admin-active-tab", tab);
	};

	return (
		<div className="h-full flex flex-col bg-background">
			{/* Tabs */}
			<div className="flex shrink-0 overflow-x-auto border-b px-6">
				{TABS.map((tab) => (
					<button
						key={tab.id}
						className={`px-4 py-2.5 text-sm font-medium transition-colors cursor-pointer ${
							activeTab === tab.id
								? "border-b-2 border-primary text-foreground"
								: "text-muted-foreground hover:text-foreground"
						}`}
						onClick={() => handleTabChange(tab.id)}
					>
						{t(tab.labelKey)}
					</button>
				))}
			</div>

			{/* Content */}
			<div className="flex-1 overflow-y-auto px-6 py-3">
				{activeTab === "overview" && <OverviewTab />}
				{activeTab === "words" && <BannedWordsTab />}
				{activeTab === "reports" && <ReportsTab />}
				{activeTab === "banned" && <BannedUsersTab />}
			</div>
		</div>
	);
}
