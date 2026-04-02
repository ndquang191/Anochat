"use client";

import { useState } from "react";
import { BannedWordsTab } from "./banned-words-tab";
import { ReportsTab } from "./reports-tab";
import { BannedUsersTab } from "./banned-users-tab";

type Tab = "words" | "reports" | "banned";

const TABS: { id: Tab; label: string }[] = [
	{ id: "reports", label: "Reports" },
	{ id: "banned", label: "Banned Users" },
	{ id: "words", label: "Banned Words" },
];

export function AdminPanel() {
	const [activeTab, setActiveTab] = useState<Tab>(() => {
		if (typeof window !== "undefined") {
			return (localStorage.getItem("admin_active_tab") as Tab) || "reports";
		}
		return "reports";
	});

	const handleTabChange = (tab: Tab) => {
		setActiveTab(tab);
		localStorage.setItem("admin_active_tab", tab);
	};

	return (
		<div className="h-full flex flex-col bg-background">
			{/* Tabs */}
			<div className="flex border-b px-6 shrink-0">
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
						{tab.label}
					</button>
				))}
			</div>

			{/* Content */}
			<div className="flex-1 overflow-y-auto px-6 py-4">
				{activeTab === "words" && <BannedWordsTab />}
				{activeTab === "reports" && <ReportsTab />}
				{activeTab === "banned" && <BannedUsersTab />}
			</div>
		</div>
	);
}
