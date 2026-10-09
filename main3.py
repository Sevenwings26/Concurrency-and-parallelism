"""
Python Concurrency Examples (Comparing to Go's Worker Pool)
"""

import asyncio
import time
from concurrent.futures import ThreadPoolExecutor

# ==========================================
# 1. ThreadPoolExecutor (Multi-threading)
# ==========================================
# Similar to your original Go Worker Pool but uses OS Threads.
# Hardware Impact: High memory usage per thread (~1-8MB). Heavy OS context switching.
# Not truly parallel for CPU-bound tasks due to the GIL. Great for legacy I/O tasks.

def process_job(job_id):
    print(f"[Thread Pool] Started job {job_id}")
    time.sleep(1) # Simulating I/O
    print(f"[Thread Pool] Finished job {job_id}")
    return True

def run_thread_pool():
    print("--- Running Python ThreadPool (Like basic Go workers) ---")
    # This automatically acts like a bounded worker pool
    with ThreadPoolExecutor(max_workers=3) as executor:
        for i in range(1, 6):
            executor.submit(process_job, i)


# ==========================================
# 2. AsyncIO (Modern Python Concurrency)
# ==========================================
# Used by FastAPI. Single-threaded, cooperative multitasking.
# Hardware Impact: Very low memory usage. Fast context switching within the Python process.
# Danger: A single CPU-bound blocking call (e.g., time.sleep instead of asyncio.sleep) 
# will freeze ALL workers and the event loop.

async def async_worker(worker_id, queue: asyncio.Queue, cancel_event: asyncio.Event):
    while not cancel_event.is_set():
        try:
            # Wait for a job, but respect cancellation
            job_id = await asyncio.wait_for(queue.get(), timeout=0.5)
        except asyncio.TimeoutError:
            continue # Check cancel_event again
            
        if job_id is None: # Sentinel value to shutdown
            queue.task_done()
            break
            
        print(f"[Async Worker {worker_id}] Started job {job_id}")
        
        # Simulating Async I/O (like an async DB call). 
        # MUST use await, otherwise the whole loop blocks!
        await asyncio.sleep(1) 
        
        print(f"[Async Worker {worker_id}] Finished job {job_id}")
        queue.task_done()

async def run_async_pool():
    print("\n--- Running Python AsyncIO (With Graceful Shutdown & Backpressure) ---")
    # Bounded queue creates backpressure, similar to Go's buffered channels
    queue = asyncio.Queue(maxsize=5) 
    cancel_event = asyncio.Event()
    
    # Start 3 workers
    workers = []
    for i in range(1, 4):
        task = asyncio.create_task(async_worker(i, queue, cancel_event))
        workers.append(task)
        
    # Produce jobs
    for i in range(1, 6):
        await queue.put(i)
        print(f"[Async Producer] Queued job {i}")
        
    # Wait for queue to process
    await queue.join()
    
    # Graceful shutdown (send sentinels)
    for _ in range(3):
        await queue.put(None)
        
    await asyncio.gather(*workers)
    print("All async jobs completed gracefully.")

if __name__ == "__main__":
    run_thread_pool()
    asyncio.run(run_async_pool())
