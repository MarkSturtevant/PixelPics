import PocketBase, { ClientResponseError, RecordService } from 'pocketbase';

interface RecordModel {
	id: string;
	created: string;
	updated: string;
	collectionId: string;
	collectionName: string;
}

export const pb = new PocketBase(window.location.origin);

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
				t: 'set';
				r: number;
				c: number;
		  }
		| {
				t: 'unset';
				r: number;
				c: number;
		  }
		| {
				t: 'mark';
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
	avatar: string;
};

export const userCollection = pb.collection('users') as RecordService<UserRecord>;

export async function logInWithGoogle(): Promise<'success' | 'failed'> {
	const authData = await pb.collection('users').authWithOAuth2({ provider: 'google' });
	if (!pb.authStore.isValid) return 'failed';
	const meta = authData.meta;
	if (!meta) return 'success';
	if (meta.isNew) {
		const formData = new FormData();
		const response = await fetch(meta.avatarUrl);
		if (response.ok) {
			const file = await response.blob();
			formData.append('avatar', file);
		}
		formData.append('name', meta.name);
		await pb.collection('users').update(authData.record.id, formData);
	}
	return 'success';
}

export async function signUpWithUsername(
	email: string,
	password: string,
	name: string
): Promise<'success' | 'empty-email' | 'empty-password' | 'invalid-credentials'> {
	if (!email) return 'empty-email';
	if (!password) return 'empty-password';
	const users = await userCollection.getList(1, 1, { filter: `email = '${email}'` });
	if (users.items.length === 0) {
		try {
			await pb.collection('users').create({ email, password, passwordConfirm: password, name });
		} catch (e: unknown) {
			const err = e as ClientResponseError;
			void err;
			return 'invalid-credentials';
		}
	}
	try {
		await pb.collection('users').authWithPassword(email, password);
	} catch (e: unknown) {
		const err = e as ClientResponseError;
		void err;
		return 'invalid-credentials';
	}
	if (!pb.authStore.isValid) return 'invalid-credentials';
	return 'success';
}

export async function logInWithUsername(
	email: string,
	password: string
): Promise<'success' | 'empty-email' | 'empty-password' | 'invalid-credentials'> {
	if (!email) return 'empty-email';
	if (!password) return 'empty-password';
	const users = await userCollection.getList(1, 1, { filter: `email = '${email}'` });
	if (users.items.length === 0) return 'invalid-credentials';
	try {
		await pb.collection('users').authWithPassword(email, password);
	} catch (e: unknown) {
		const err = e as ClientResponseError;
		void err;
		return 'invalid-credentials';
	}
	if (!pb.authStore.isValid) return 'invalid-credentials';
	return 'success';
}
