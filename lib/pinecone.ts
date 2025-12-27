import { Pinecone } from '@pinecone-database/pinecone';
import { PineconeQueryResult, PartnerShop } from './types';

let pineconeClient: Pinecone | null = null;

function getClient(): Pinecone {
  if (!pineconeClient) {
    const apiKey = process.env.PINECONE_API_KEY;
    if (!apiKey) {
      throw new Error('PINECONE_API_KEY is not configured');
    }
    pineconeClient = new Pinecone({ apiKey });
  }
  return pineconeClient;
}

// OpenAI Embedding API를 사용한 벡터 생성
async function createEmbedding(text: string): Promise<number[]> {
  const apiKey = process.env.OPENROUTER_API_KEY;
  if (!apiKey) {
    throw new Error('OPENROUTER_API_KEY is not configured for embeddings');
  }

  // OpenAI embedding endpoint 사용 (OpenRouter는 embedding을 지원하지 않음)
  // 대안: text-embedding-3-small을 직접 호출하거나, Pinecone의 자체 embedding 사용
  const response = await fetch('https://api.openai.com/v1/embeddings', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model: 'text-embedding-3-small',
      input: text,
    }),
  });

  if (!response.ok) {
    // Embedding 실패시 빈 배열 반환 (graceful degradation)
    console.error('Embedding API failed, skipping vector search');
    return [];
  }

  const data = await response.json();
  return data.data[0].embedding;
}

// Knowledge Index 검색
export async function searchLaundryKnowledge(
  query: string,
  topK: number = 3
): Promise<PineconeQueryResult[]> {
  try {
    const indexName = process.env.PINECONE_INDEX_LAUNDRY;
    if (!indexName) {
      console.warn('PINECONE_INDEX_LAUNDRY not configured');
      return [];
    }

    const embedding = await createEmbedding(query);
    if (embedding.length === 0) {
      return [];
    }

    const client = getClient();
    const index = client.index(indexName);

    const results = await index.query({
      vector: embedding,
      topK,
      includeMetadata: true,
    });

    return (results.matches || []).map((match) => ({
      id: match.id,
      score: match.score || 0,
      metadata: match.metadata || {},
    }));
  } catch (error) {
    console.error('Error searching laundry knowledge:', error);
    return [];
  }
}

// Partner Index 검색
export async function searchPartnerShops(
  query: string,
  zipcode?: string,
  topK: number = 3
): Promise<PartnerShop[]> {
  try {
    const indexName = process.env.PINECONE_INDEX_PARTNER;
    if (!indexName) {
      console.warn('PINECONE_INDEX_PARTNER not configured');
      return [];
    }

    const embedding = await createEmbedding(query);
    if (embedding.length === 0) {
      return [];
    }

    const client = getClient();
    const index = client.index(indexName);

    // 필터 조건: active subscription + zipcode (선택적)
    const filter: Record<string, unknown> = {
      subscription: 'active',
    };
    if (zipcode) {
      filter.zipcode = zipcode;
    }

    const results = await index.query({
      vector: embedding,
      topK,
      includeMetadata: true,
      filter,
    });

    return (results.matches || []).map((match) => ({
      shop_name: (match.metadata?.shop_name as string) || 'Unknown Shop',
      zipcode: (match.metadata?.zipcode as string) || '',
      subscription: (match.metadata?.subscription as 'active' | 'inactive') || 'active',
      specialty: (match.metadata?.specialty as string[]) || [],
      rating: match.metadata?.rating as number | undefined,
    }));
  } catch (error) {
    console.error('Error searching partner shops:', error);
    return [];
  }
}

// Knowledge 검색 결과를 Context 문자열로 변환
export function formatKnowledgeAsContext(results: PineconeQueryResult[]): string {
  if (results.length === 0) {
    return '';
  }

  return results
    .map((result, index) => {
      const meta = result.metadata;
      const title = meta.title || meta.stain_type || 'Unknown';
      const content = meta.content || meta.description || '';
      const successRate = meta.success_rate ? `(성공률: ${meta.success_rate})` : '';

      return `${index + 1}. ${title} ${successRate}\n${content}`;
    })
    .join('\n\n');
}
