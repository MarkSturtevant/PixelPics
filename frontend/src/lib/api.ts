import PocketBase, { RecordService } from 'pocketbase';

interface RecordModel {
	id: string;
	created: string;
	updated: string;
	collectionId: string;
	collectionName: string;
}

export const pb = new PocketBase("/");

export type PuzzleRecord = RecordModel & {
	name: string;
	background_image: string;
	puzzle_image: string;
	puzzle: {
		pd: string;
		rl: number;
		cl: number;
	};
	style_meta: {
		'background-color': string;
		'puzzle-color': string;
	};
	category: string;
	author: string;
	expand?: {
		author?: UserRecord;
	};
};

export const puzzleCollection = pb.collection('puzzles') as RecordService<PuzzleRecord>;

export type GameRecord = RecordModel & {
	puzzle: string;
	expand?: {
		puzzle?: PuzzleRecord;
		game_events_via_game?: GameEventRecord[];
	};
};

export const gameCollection = pb.collection('games') as RecordService<GameRecord>;

export type GameEventRecord = RecordModel & {
	game: string;
	author: string;
	action:
		| {
				t: 'S';
				r: number;
				c: number;
		  }
		| {
				t: 'U';
				r: number;
				c: number;
		  }
		| {
				t: 'M';
				r: number;
				c: number;
		  }
		| {
				t: 'X';
				r: number;
				c: number;
		  };
	expand?: Record<string, never>;
};

export const gameEventCollection = pb.collection('game_events') as RecordService<GameEventRecord>;

export type UserRecord = RecordModel & {
	email: string;
	emailVisibility: boolean;
	verified: boolean;
	name: string;
	username: string;
	avatar: string;
};

export const userCollection = pb.collection('users') as RecordService<UserRecord>;

export async function logInWithDiscord(): Promise<'success' | 'failed'> {
	await pb.collection('users').authWithOAuth2({ provider: 'discord' });
	if (!pb.authStore.isValid) return 'failed';
	return 'success';
}
