import { cookies } from "next/headers";

// Server-side cookie utilities
export const getServerCookie = async (name: string): Promise<string | undefined> => {
	const cookieStore = await cookies();
	return cookieStore.get(name)?.value;
};
