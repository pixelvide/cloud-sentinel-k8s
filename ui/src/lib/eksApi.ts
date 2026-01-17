
import { api } from "./api";

export interface EKSCluster {
    name: string;
    region: string;
    account_id: string;
}

export interface ListEKSClustersParams {
    aws_config_id: number;
    region?: string;
}

export interface ImportEKSClustersParams {
    aws_config_id: number;
    region?: string;
    cluster_names: string[];
}

export const eksApi = {
    listClusters: async (params: ListEKSClustersParams): Promise<{ clusters: EKSCluster[] }> => {
        const response = await api.post<{ clusters: EKSCluster[] }>("/settings/aws/eks/clusters", params);
        return response;
    },
    importClusters: async (params: ImportEKSClustersParams): Promise<{ message: string; imported_count: number }> => {
        const response = await api.post<{ message: string; imported_count: number }>("/settings/aws/eks/import", params);
        return response;
    },
};
