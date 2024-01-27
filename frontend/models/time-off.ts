export interface TimeOff {
    personId: number;
    category: 'vacation' | 'sickness' | 'other';
    hours: number;
}
