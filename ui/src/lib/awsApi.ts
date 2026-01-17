
import { api } from "./api";

export interface AWSConfig {
    id: number;
    user_id: number;
    name: string;
    access_key_id: string;
    region: string;
    created_at: string;
    updated_at: string;
}

export interface CreateAWSConfigParams {
    name: string;
    access_key_id: string;
    secret_access_key: string;
    session_token?: string;
    region: string;
}

export const awsApi = {
    listConfigs: async (): Promise<{ configs: AWSConfig[] }> => {
        const response = await api.get<{ configs: AWSConfig[] }>("/settings/aws");
        return response;
    },

    createConfig: async (params: CreateAWSConfigParams): Promise<AWSConfig> => {
        const response = await api.post<AWSConfig>("/settings/aws", params);
        return response;
    },

    deleteConfig: async (id: number): Promise<void> => {
        await api.del(`/settings/aws/${id}`);
    },
};
