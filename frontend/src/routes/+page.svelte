<script lang="ts">
	import { scale } from 'svelte/transition';
	import { encode, decode } from 'cbor2';
	import { nanoid } from 'nanoid';
	import { cn } from '$lib/utils';
	import { onMount, untrack } from 'svelte';
	import {gameCollection, gameEventCollection, pb} from '$lib/api';
	import { logInWithDiscord } from '$lib/api';

	pb.autoCancellation(false);

	const puzzle: ('S' | 'U' | 'M' | 'X')[][] = $state([['U']]);
	const rows: number[][] = $state([[1]]);
	const cols: number[][] = $state([[1]]);
	let puzzleSolution: string = '1';
	let puzzleColor = $state("#c084fc");
	let puzzleFinished = $state(false);

	let remRows = 141241424, remCols = 1123234125;
	function checkSolution(i: number, j: number): void {
		const N = rows.length, M = cols.length;

		// check the rows
		const arr2: number[] = [];
		let l = 0;
		for (let k = 0; k < M; ++k) {
			if (puzzle[i][k] === 'S')
				l++;
			else if (l > 0) {
				arr2.push(l);
				l = 0;
			}
		}
		if (l > 0 || arr2.length === 0)
			arr2.push(l);

		// check the cols
		const arr3: number[] = [];
		l = 0;
		for (let k = 0; k < N; ++k) {
			if (puzzle[k][j] === 'S')
				l++;
			else if (l > 0) {
				arr3.push(l);
				l = 0;
			}
		}
		if (l > 0 || arr3.length === 0)
			arr3.push(l);

		// update the remain variables
		if (rows[i].join(',') === arr2.join(','))
			remRows &= (1 << N) - 1 - (1 << i);
		else
			remRows |= (1 << i);

		if (cols[j].join(',') === arr3.join(','))
			remCols &= (1 << M) - 1 - (1 << j);
		else
			remCols |= (1 << j);


		if (remRows === 0 && remCols === 0) {
			// @ts-ignore solution is found!
			document.getElementById('puzzle-reveal').style.display = "block";
			puzzleFinished = true;
		}
	}

	let userURL: string = 'ad1bjp5ogg316z8';
	let gameURL: string = '9a951768vgnpx5v';

	onMount(() => {
		const gamePromise = gameCollection.getOne(gameURL, { expand: 'puzzle' });
		gamePromise.then((game) => {
			const data = game?.expand?.puzzle;
			if (!data) {
				console.log('jimmy is stupid');
				return;
			} // #84aff5

			// set the background colors
			document.body.style.backgroundColor = data.style_meta['background-color'];
			puzzleColor = data.style_meta['puzzle-color'];

			// @ts-ignore set the background image
			document.getElementById('body-bg-image').style.backgroundImage = `url(${pb.files.getURL(data, data.background_image)})`;

			// @ts-ignore set the puzzle image
			document.getElementById('puzzle-reveal').style.backgroundImage = `url(${pb.files.getURL(data, data.puzzle_image)})`;

			// set the correctRows / Cols variables
			remRows = (1 << data.puzzle.rl) - 1;
			remCols = (1 << data.puzzle.cl) - 1;

			// generate an empty puzzle
			puzzle.pop();
			for (let i = 0; i < data.puzzle.rl; ++i) {
				const temp: ('S' | 'U' | 'M' | 'X')[] = [];
				for (let j = 0; j < data.puzzle.cl; ++j) temp.push('U');
				puzzle.push(temp);
			}

			// get the puzzle solution
			puzzleSolution = data.puzzle.pd;

			// parse the current puzzle progress
			const eventsPromise = gameEventCollection.getFullList({
				filter: 'game = "' + gameURL + '"',
				sort: 'created'
			});
			eventsPromise.then((eventsList) => {
				for (let event of eventsList) {
					puzzle[event.action.r][event.action.c] = event.action.t;
					checkSolution(event.action.r, event.action.c);
				}
			});

			// parse the rows
			rows.pop();
			for (let i = 0; i < data.puzzle.rl; ++i) {
				const temp: number[] = [];
				let curCnt = 0;
				for (let j = 0; j < data.puzzle.cl; ++j) {
					const p = i * data.puzzle.cl + j;
					if (puzzleSolution[p] === '1') curCnt++;
					if (curCnt > 0 && (puzzleSolution[p] === '0' || j === data.puzzle.cl - 1)) {
						temp.push(curCnt);
						curCnt = 0;
					}
				}
				if (temp.length === 0) {
					temp.push(0);
					remRows -= (1 << i);
				}
				rows.push(temp);
			}
			if (rows.length === 0) rows.push([0]);

			// parse the cols
			cols.pop();
			for (let i = 0; i < data.puzzle.cl; ++i) {
				const temp: number[] = [];
				let curCnt = 0;
				for (let j = 0; j < data.puzzle.rl; ++j) {
					const p = j * data.puzzle.cl + i;
					if (puzzleSolution[p] === '1') curCnt++;
					if (curCnt > 0 && (puzzleSolution[p] === '0' || j === data.puzzle.rl - 1)) {
						temp.push(curCnt);
						curCnt = 0;
					}
				}
				if (temp.length === 0) {
					temp.push(0);
					remCols -= (1 << i);
				}
				cols.push(temp);
			}
		});

		gameEventCollection.subscribe('*', function (data) {
			console.log(data);
			puzzle[data.record.action.r][data.record.action.c] = data.record.action.t;
		});
	});

	let mouse_down: number = -1;
	let starter = 'U';
	let session: Set<number> = new Set();

	function ev_mouse_down(i: number, j: number, e: MouseEvent): void {
		mouse_down = e.button;
		starter = puzzle[i][j];
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
		if (puzzle[i][j] != starter) return;

		if (session.has(i * 10000 + j)) return;
		session.add(i * 10000 + j);

		if (typ === 'S') {
			// press S: solid tile
			if (puzzle[i][j] == 'X') return;
			puzzle[i][j] = puzzle[i][j] === 'S' ? 'U' : 'S';
		} else if (typ === 'X') {
			// press X: x tile
			if (puzzle[i][j] == 'S') return;
			puzzle[i][j] = puzzle[i][j] === 'X' ? 'U' : 'X';
		} else if (typ === 'M') {
			// press M: mark tile
			if (puzzle[i][j] === 'S' || puzzle[i][j] === 'X') return;
			puzzle[i][j] = puzzle[i][j] === 'M' ? 'U' : 'M';
		}

		// create a game event and publish it
		const data = {
			game: gameURL,
			author: userURL,
			action: {
				t: puzzle[i][j],
				r: i,
				c: j
			}
		};
		gameEventCollection.create(data);

		checkSolution(i, j);
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

<button
	onclick={() => {
		logInWithDiscord();
	}}
>
	Discord
</button>

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

		<div class="relative">
			<div
				class="grid border border-white absolute"
				style:grid-template-rows="repeat({puzzle.length}, minmax(0, 1fr))"
				style:grid-template-columns="repeat({puzzle[0].length}, minmax(0, 1fr))"
			>
				{#each puzzle as row, i}
					{#each row as cell, j}
						<div
							class={cn(
								'aspect-square size-10 place-content-center place-items-center outline-none',
								'border-white hover:border-amber-400 hover:border-2',
								i !== 0 && 'border-t-[0.5px]',
								j !== 0 && 'border-l-[0.5px]',
								i !== puzzle.length - 1 && 'border-b-[0.5px]',
								j !== puzzle.length - 1 && 'border-r-[0.5px]',
								i % 5 === 0 && 'border-t-amber-500',
								j % 5 === 0 && 'border-l-amber-500',
								i % 5 === 4 && 'border-b-amber-500',
								j % 5 === 4 && 'border-r-amber-500'
							)}
							onmouseenter={() => test_trigger(i, j)}
							onmousedown={(e) => ev_mouse_down(i, j, e)}
							onmouseup={() => ev_mouse_up()}
							role="button"
							tabindex="0"
							style:background-color={puzzleColor}
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
			<div id="puzzle-reveal" class="object-fit border-amber-400 ease-linear border-4 bg-cover resize z-10 w-full h-full absolute transition-opacity"
				 style:image-rendering="pixelated" style:display="none" style:opacity={puzzleFinished ? 1 : 0} style:transition-duration="1500ms"
			></div>
		</div>
	</div>
</div>
{#each Object.keys(liveCursors) as key}
	<div class="absolute" style:top="{liveCursors[key].y}px" style:left="{liveCursors[key].x}px">
		<div class="h-4 w-4 rounded-full bg-red-500"></div>
	</div>
{/each}
