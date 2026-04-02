"use client";

import React, { createContext, useContext, useState } from "react";

interface AdminContextValue {
	isAdminOpen: boolean;
	setIsAdminOpen: (open: boolean) => void;
}

export const AdminContext = createContext<AdminContextValue>({
	isAdminOpen: false,
	setIsAdminOpen: () => {},
});

export function AdminProvider({ children }: { children: React.ReactNode }) {
	const [isAdminOpen, setIsAdminOpen] = useState(false);
	return (
		<AdminContext.Provider value={{ isAdminOpen, setIsAdminOpen }}>
			{children}
		</AdminContext.Provider>
	);
}

export const useAdmin = () => useContext(AdminContext);
