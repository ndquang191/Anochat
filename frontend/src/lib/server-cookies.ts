import { cookies } from "next/headers";

// Server-side cookie utilities
export const getServerCookie = (name: string): string | undefined => {
	const cookieStore = cookies();
	return cookieStore.get(name)?.value;
};
