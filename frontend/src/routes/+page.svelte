<script lang="ts">
	import { scale } from 'svelte/transition';
	import { encode, decode } from 'cbor2';
	import { nanoid } from 'nanoid';
	import { cn } from '$lib/utils';
	import { untrack } from 'svelte';

	const rawPuzzle: string[] = [
		'XXXXXXXXXXXXXXX',
		'XXXXXXXXXXXXXXX',
		'XSXXXXXXXXXXXXX',
		'SSSSXXXXXXXXXXX',
		'SXSSSSXXXXXXXXX',
		'SXSSXSSSSXXXXXX',
		'SXSSSSSXSSSSSSX',
		'SXSSSXSSSSXSXSS',
		'XSXSSSSSXSSSXSS',
		'XSSXSSSSSSSXSSS',
		'XXSSXSSSSSXSSSX',
		'XXXSSXXXXXSSSXX',
		'XXXXSSSSSSSXXXX',
		'XXXXXXXXXXXXXXX',
		'XXXXXXXXXXXXXXX'
	];

	function convertRawToPuzzle(raw: string[]) {
		return raw.map((row) => row.split('') as ('S' | 'U' | 'M' | 'X')[]);
	}

	const puzzle: ('S' | 'U' | 'M' | 'X')[][] = $state(convertRawToPuzzle(rawPuzzle));

	const rows: number[][] = $state([
		[0],
		[0],
		[1],
		[4],
		[1, 4],
		[1, 2, 4],
		[1, 5, 6],
		[1, 3, 4, 1, 2],
		[1, 5, 3, 2],
		[2, 7, 3],
		[2, 5, 3],
		[2, 3],
		[7],
		[0],
		[0]
	]);

	const cols: number[][] = $state([
		[5],
		[2, 2],
		[5, 2],
		[6, 2],
		[1, 4, 2],
		[3, 3, 1],
		[6, 1],
		[1, 4, 1],
		[3, 2, 1],
		[5, 1],
		[1, 2, 2],
		[3, 2],
		[1, 3],
		[5],
		[3]
	]);

	let mouse_down: number = -1;
	let session: Set<number> = new Set();

	function ev_mouse_down(i: number, j: number, e: MouseEvent): void {
		mouse_down = e.button;
		console.log('hiw');
		test_trigger(i, j);
	}
	function ev_mouse_up(): void {
		mouse_down = -1;
		session.clear();
	}

	function test_trigger(i: number, j: number): void {
		if (mouse_down === 0) flip(i, j, 'S');
		else if (mouse_down === 2) flip(i, j, 'X');
		else if (mouse_down === 1) flip(i, j, 'M');
	}

	function flip(i: number, j: number, typ: string): void {
		if (i == -1 || j == -1) return;

		if (session.has(i * 10000 + j)) return;
		session.add(i * 10000 + j);

		if (typ === 'S') {
			// press S
			if (puzzle[i][j] == 'X') return;
			puzzle[i][j] = puzzle[i][j] === 'S' ? 'U' : 'S';
		} else if (typ === 'X') {
			// press X
			if (puzzle[i][j] == 'S') return;
			puzzle[i][j] = puzzle[i][j] === 'X' ? 'U' : 'X';
		} else if (typ === 'M') {
			if (puzzle[i][j] === 'S' || puzzle[i][j] === 'X') return;
			puzzle[i][j] = puzzle[i][j] === 'M' ? 'U' : 'M';
		}
	}

	const clientID = nanoid();

	let socket: WebSocket;

	const wsRequestMessageAwareness = 1;
	const wsResponseMessageAwareness = 1;
	type WSRequestMessageType = typeof wsRequestMessageAwareness;
	type WSResponseMessageType = typeof wsResponseMessageAwareness;

	const awareness = $state<Record<string, Record<string, unknown>>>({});
	const localAwareness = $state<Record<string, unknown>>({});

	function sendWSMessage(typ: WSRequestMessageType, msg: unknown): void {
		if (socket.readyState !== WebSocket.OPEN) return;
		const data = encode({ t: typ, m: msg });
		socket.send(data);
	}

	$effect(() => {
		untrack(() => {
			socket = new WebSocket(`/api/v1/ws?id=${clientID}`);
			socket.binaryType = 'arraybuffer';
			socket.onmessage = (event) => {
				const data = decode(new Uint8Array(event.data)) as { t: WSResponseMessageType; m: unknown };
				switch (data.t) {
					case wsResponseMessageAwareness: {
						const msg = data.m as Record<string, Record<string, unknown>>;
						for (const key in msg) {
							if (key === clientID) continue;
							if (msg[key] === null) delete awareness[key];
							else awareness[key] = msg[key];
						}
						break;
					}
				}
			};
			socket.onclose = () => {
				console.log('Socket is closed.');
			};
			socket.onerror = (err) => {
				console.error('Socket encountered error: ', err);
				socket.close();
			};
			socket.onopen = () => {
				console.log('Connected to server');
			};
		});
	});

	$effect(() => {
		if (Object.keys(localAwareness).length === 0) return;
		sendWSMessage(wsRequestMessageAwareness, localAwareness);
	});

	let puzzleDiv = $state<HTMLElement | null>(null);

	function updateCursor(clientX: number, clientY: number): void {
		if (!puzzleDiv) return;
		const rect = puzzleDiv.getBoundingClientRect();
		const yRelativeCoord = (clientY - rect.top) / rect.height;
		const xRelativeCoord = (clientX - rect.left) / rect.width;
		localAwareness.cursorX = xRelativeCoord;
		localAwareness.cursorY = yRelativeCoord;
	}

	const liveCursors = $derived.by<Record<string, { x: number; y: number }>>(() => {
		if (!puzzleDiv) return {};
		const rect = puzzleDiv.getBoundingClientRect();
		const cursors: Record<string, { x: number; y: number }> = {};
		for (const key in awareness) {
			const cursor = awareness[key];
			if (typeof cursor.cursorX !== 'number' || typeof cursor.cursorY !== 'number') continue;
			const xAbsCoord = Math.max(
				0,
				Math.min(cursor.cursorX * rect.width + rect.left, window.innerWidth)
			);
			const yAbsCoord = Math.max(
				0,
				Math.min(cursor.cursorY * rect.height + rect.top, window.innerHeight)
			);
			cursors[key] = { x: xAbsCoord, y: yAbsCoord };
		}
		return cursors;
	});

	// const TILE_SIZE: number = $state(40);
	// const MINI_SIZE: number = $derived(TILE_SIZE * 0.8);
</script>

<div
	class="flex h-screen w-screen items-center justify-center"
	role="main"
	onmousemove={(e) => {
		updateCursor(e.clientX, e.clientY);
	}}
>
	<div bind:this={puzzleDiv} class="grid grid-cols-[max-content_1fr] grid-rows-[max-content_1fr]">
		<div></div>

		<div class="flex flex-row" style:padding-left="1px" style:padding-right="1px">
			{#each cols as col, i}
				<div
					class={cn(
						'flex w-10 flex-col items-center justify-end border-transparent bg-gradient-to-b from-transparent pb-2',
						i % 2 === 0 ? 'to-black/20' : 'to-white/20',
						i !== 0 && 'border-l-[0.5px]',
						i !== cols.length - 1 && 'border-r-[0.5px]'
					)}
					style:gap="12px"
				>
					{#each col as num}
						<div>
							<span
								class="font-sans font-bold text-black drop-shadow-[0_1.5px_1.5px_rgba(255,255,255,0.8)]"
								style:font-size="24px"
								style:line-height="24px">{num}</span
							>
						</div>
					{/each}
				</div>
			{/each}
		</div>

		<div class="flex w-max flex-col" style:padding-top="1px" style:padding-bottom="1px">
			{#each rows as row, i}
				<div
					class={cn(
						'flex h-10 flex-row items-center justify-end border-transparent bg-gradient-to-r from-transparent pr-2',
						i % 2 === 0 ? 'to-black/20' : 'to-white/20',
						i !== 0 && 'border-t-[0.5px]',
						i !== rows.length - 1 && 'border-b-[0.5px]'
					)}
					style:gap="12px"
				>
					{#each row as num}
						<div style:width="24px" class="flex flex-row items-center justify-center">
							<span
								class="font-sans font-bold text-black drop-shadow-[0_1.5px_1.5px_rgba(255,255,255,0.8)]"
								style:font-size="24px"
								style:line-height="24px"
							>
								{num}
							</span>
						</div>
					{/each}
				</div>
			{/each}
		</div>

		<div
			class="grid border border-white"
			style:grid-template-rows="repeat({puzzle.length}, minmax(0, 1fr))"
			style:grid-template-columns="repeat({puzzle[0].length}, minmax(0, 1fr))"
		>
			{#each puzzle as row, i}
				{#each row as cell, j}
					<div
						class={cn(
							'aspect-square size-10 place-content-center place-items-center bg-purple-400 outline-none',
							'border-purple-200 hover:border-amber-400',
							i !== 0 && 'border-t-[0.5px]',
							j !== 0 && 'border-l-[0.5px]',
							i !== puzzle.length - 1 && 'border-b-[0.5px]',
							j !== puzzle.length - 1 && 'border-r-[0.5px]'
						)}
						onmouseenter={() => test_trigger(i, j)}
						onmousedown={(e) => ev_mouse_down(i, j, e)}
						onmouseup={() => ev_mouse_up()}
						role="button"
						tabindex="0"
					>
						{#if cell === 'S'}
							<div transition:scale class="aspect-square rounded bg-white" style:width="32px"></div>
						{:else if cell === 'U'}
							<div transition:scale></div>
						{:else if cell === 'M'}
							<div
								transition:scale
								class="aspect-square size-8 rounded border-2 border-white"
							></div>
						{:else if cell === 'X'}
							<div transition:scale class="relative aspect-square size-10">
								<div class="absolute aspect-square size-10">
									<div class="flex h-full w-full items-center justify-center">
										<div class="h-8 w-0.5 rotate-45 rounded-full bg-black"></div>
									</div>
								</div>
								<div class="absolute aspect-square size-10">
									<div class="flex h-full w-full items-center justify-center">
										<div class="h-8 w-0.5 -rotate-45 rounded-full bg-black"></div>
									</div>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			{/each}
		</div>
	</div>
</div>
{#each Object.keys(liveCursors) as key}
	<div class="absolute" style:top="{liveCursors[key].y}px" style:left="{liveCursors[key].x}px">
		<div class="h-4 w-4 rounded-full bg-red-500"></div>
	</div>
{/each}
