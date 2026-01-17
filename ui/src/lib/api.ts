import { API_URL } from "./config";

export async function apiFetch(path: string, options: RequestInit = {}) {
    // If it's a relative path, prepend API_URL
    const url = path.startsWith("http") || path.startsWith("/") && path.startsWith(API_URL)
        ? path
        : `${API_URL}${path.startsWith("/") ? "" : "/"}${path}`;

    const defaultOptions: RequestInit = {
        credentials: "include",
        ...options,
        headers: {
            "Content-Type": "application/json",
            ...options.headers,
        },
    };

    if (options.body instanceof FormData) {
        // Let the browser set the Content-Type with the boundary
        delete (defaultOptions.headers as any)["Content-Type"];
    }

    const response = await fetch(url, defaultOptions);

    if (response.status === 401) {
        if (typeof window !== "undefined") {
            window.location.href = "/login";
        }
    }

    return response;
}

export async function get<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, { ...options, method: "GET" });
    if (!response.ok) {
        throw new Error(`API Error: ${response.statusText}`);
    }
    return response.json();
}

export async function post<T>(path: string, body?: any, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, {
        ...options,
        method: "POST",
        body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
        throw new Error(`API Error: ${response.statusText}`);
    }
    return response.json();
}

/**
 * Constructs a WebSocket URL for a given path, using the API_URL as a base if it's external.
 * @param path The API path (e.g., "/kube/logs")
 * @returns The full WebSocket URL
 */
export function getWsUrl(path: string): string {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    let wsHost = window.location.host;

    // If API_URL is an absolute URL, extract the host
    if (API_URL.startsWith("http")) {
        try {
            const url = new URL(API_URL);
            wsHost = url.host;
        } catch (e) {
            console.error("Invalid API_URL for WebSocket construction:", API_URL);
        }
    }

    const cleanPath = path.startsWith("/") ? path : `/${path}`;
    return `${protocol}//${wsHost}${cleanPath}`;
}

export async function put<T>(path: string, body?: any, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, {
        ...options,
        method: "PUT",
        body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
        throw new Error(`API Error: ${response.statusText}`);
    }
    return response.json();
}

export async function del<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, { ...options, method: "DELETE" });
    if (!response.ok) {
        throw new Error(`API Error: ${response.statusText}`);
    }
    return response.json();
}

export const api = {
    get,
    post,
    put,
    del,
    fetch: apiFetch,
    getWsUrl,
    getPodMetrics: async (context: string, namespace: string, podName: string) => {
        return get<{
            cpu: Array<{ timestamp: string; value: number }>;
            memory: Array<{ timestamp: string; value: number }>;
            fallback: boolean;
        }>(`/kube/metrics/pods?namespace=${encodeURIComponent(namespace)}&podName=${encodeURIComponent(podName)}`, {
            headers: {
                "x-kube-context": context
            }
        });
    }
};
