export interface Account {
    id: number;
    accountNumber: string;
    userId: string;
    balance: number;
    equity?: number;
    margin?: number;
    freeMargin?: number;
    marginLevel?: number;
    leverage: number;
    marginMode: string;
    currency: string;
    status: string;
    isDemo: boolean;
}

export interface LedgerEntry {
    id: number;
    accountId: number;
    type: string;
    amount: number;
    balanceAfter: number;
    description: string;
    paymentMethod: string;
    paymentRef: string;
    adminId: string;
    createdAt: string;
}

export interface RoutingRule {
    id: string;
    groupPattern: string;
    symbolPattern: string;
    minVolume: number;
    maxVolume: number;
    action: string;
    targetLp: string;
    priority: number;
}
