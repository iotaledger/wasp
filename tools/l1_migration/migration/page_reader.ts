// Define the common pagination structure
export interface PaginatedResponse<T> {
  data: T[];
  hasNextPage: boolean;
  nextCursor?: string | null;
}

// Define the type for pagination request functions
export type PaginationRequestFn<T, P> = (params: P & { cursor?: string }) => Promise<PaginatedResponse<T>>;

// Generic function to handle paginated requests
export async function paginatedRequest<T, P>(requestFn: PaginationRequestFn<T, P>, params: P): Promise<T[]> {
  const allItems: T[] = [];
  let currentCursor: string | undefined;

  try {
    do {
      const result = await requestFn({
        ...params,
        cursor: currentCursor,
      });

      if (!result?.data) {
        throw new Error('Invalid response: missing data');
      }

      allItems.push(...result.data);
      currentCursor = result.hasNextPage ? (result.nextCursor ?? undefined) : undefined;
    } while (currentCursor !== undefined);

    return allItems;
  } catch (error) {
    throw new Error(`Failed to fetch paginated data: ${error.message}`);
  }
}
