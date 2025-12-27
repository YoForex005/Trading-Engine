export interface Point {
    time: number; // Unix timestamp
    price: number;
}

export interface Drawing {
    id: string;
    accountId: number;
    symbol: string;
    type: 'line' | 'horizontal_line' | 'rectangle' | 'text';
    points: Point[];
    options?: Record<string, any>;
}
